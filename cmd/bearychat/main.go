/*
 * Copyright (c) 2018 LI Zhennan
 *
 * Use of this work is governed by an MIT License.
 * You may find a license copy in project root.
 */

// QRCode API integration on Bearychat,
// a Slack-like working group.
//
// Most users speaks Chinese.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nanmu42/qrcode-api"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/nanmu42/qrcode-api/cmd/common"
	"github.com/pkg/errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/nanmu42/bearychat-go/openapi"

	"github.com/nanmu42/bearychat-go"
)

var (
	// config file location
	configFile *string
	logger     *zap.Logger
	// Version build params
	Version string
	// BuildDate build params
	BuildDate string

	// download image for scanning
	downloader = http.Client{
		Timeout: 30 * time.Second,
	}
)

const (
	srctag          = "bearychat"
	helpCommand     = "help"
	helpContent     = "用法：\n* 二维码生成\n```@小码 {内容}```\n内容左右的空格和回车会被忽略。\n* 二维码识别\n群聊中，引用已发出的图片消息并`@小码`，私聊中可直接发送图片。\n提示：私聊中，`@小码`需要省略。"
	helloCommand    = "hello"
	helloContent    = "小码来啦！驾～ []~(￣▽￣)~*"
	fileAPIEndpoint = "https://api.bearychat.com/v1/file.location?"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	configFile = flag.String("config", "config.toml", "config.toml file location for rly")
	w := common.NewBufferedLumberjack(&lumberjack.Logger{
		Filename:   "logs/qrcode-bot.log",
		MaxSize:    300, // megabytes
		MaxBackups: 5,
		MaxAge:     28, // days
	}, 32*1024)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(w),
		zap.InfoLevel,
	)
	logger = zap.New(core)
}

func main() {
	var err error
	defer logger.Sync()
	defer func() {
		if err != nil {
			fmt.Println(err)
			logger.Error("fatal error", zap.Error(err))
			os.Exit(1)
		}
	}()

	flag.Parse()
	fmt.Printf(`Bearychat QR Code Bot(%s)
built on %s

`, Version, BuildDate)

	err = C.LoadFrom(*configFile)
	if err != nil {
		err = errors.Wrap(err, "C.LoadFrom")
		return
	}

	botCtx, err := bearychat.NewRTMContext(C.RTMToken)
	if err != nil {
		err = errors.Wrap(err, "bearychat.NewRTMContext")
		return
	}

	err, messageChan, errChan := botCtx.Run()
	if err != nil {
		err = errors.Wrap(err, "botCtx.Run")
		return
	}

	botAPI := openapi.NewClient(C.RTMToken)

	var exitSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, exitSignals...)

	fmt.Println("Bearychat QR Code Bot started :)")
	logger.Info("Bearychat QR Code Bot started")

	for {
		select {
		case <-quitChan:
			fmt.Println("Bearychat QR Code Bot is exiting safely...")
			logger.Info("Bearychat QR Code Bot is exiting safely...")
			return
		case rtmErr := <-errChan:
			logger.Error("RTM error", zap.Error(rtmErr))
		case incoming := <-messageChan:
			if incoming.IsFromUID(botCtx.UID()) {
				continue
			}

			// only reply to mentioned
			if mentioned, content := incoming.ParseMentionUID(botCtx.UID()); mentioned {
				logger.Info("triggered",
					zap.Any("uid", incoming["uid"]),
					zap.Any("text", incoming["text"]),
				)

				var vChannelID string
				if incoming.Type() == bearychat.RTMMessageTypeUpdateAttachments {
					data, ok := incoming["data"].(map[string]interface{})
					if ok {
						vChannelID, _ = data["vchannel_id"].(string)
					}
				} else {
					vChannelID, _ = incoming["vchannel_id"].(string)
				}
				if len(vChannelID) == 0 {
					logger.Error("can not get vchannel_id")
					continue
				}

				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				msgOpt := openapi.MessageCreateOptions{
					VChannelID: vChannelID,
				}

				// j, _ := json.Marshal(incoming)
				// fmt.Printf("received: %s\n", j)

				if incoming.Type() == bearychat.RTMMessageTypeUpdateAttachments ||
					(incoming.Type() == bearychat.RTMMessageTypeP2PMessage && incoming.Text() == "上传了图片") {
					// decode
					file, badFile := incoming.ParseAttachedFile()
					switch true {
					case badFile != nil:
						msgOpt.Text = "没能在您引用的消息中找到图片附件，您可以复制图片单独发一下再试试。:face_with_cowboy_hat:  \n" + helpContent
					case file.Size >= C.MaxDecodeFileSize:
						msgOpt.Text = "图片附件体积过大，抱歉。:ghost: "
					case file.MIME != "image/jpeg" && file.MIME != "image/png" && file.MIME != "image/gif":
						msgOpt.Text = "仅支持jpeg, png或gif格式的图片哟。:kissing_heart: "
					default:
						scanningResult, badScan := DownloadImageAndScan(file.ImageURL)
						if badScan != nil {
							msgOpt.Text = "哦噢，出错了。:dizzy_face: " + badScan.Error()
						} else if len(scanningResult) == 0 {
							msgOpt.Text = "没能在您的图片中找到二维码/条形码，或者它们损坏了，小码会继续努力哒！ :kissing_heart: "
						} else {
							msgOpt.Text = ":sunglasses:  扫描结果如下：\n" + strings.Join(scanningResult, "\n")
						}
					}
					outgoing, _, badLuck := botAPI.Message.Create(ctx, &msgOpt)
					cancel()
					if badLuck != nil {
						logger.Error("failed to make response", zap.Any("outgoing msg", outgoing), zap.Error(badLuck))
					}
					continue
				}

				if referKeyInterface, referring := incoming["refer_key"]; referring {
					if referKey, ok := referKeyInterface.(string); ok && len(referKey) > 0 {
						continue
					}
				}

				content = strings.Trim(content, " \n\r\t")
				// encode or simple command
				switch true {
				case strings.ToLower(content) == helpCommand:
					msgOpt.Text = helpContent
				case strings.ToLower(content) == helloCommand:
					msgOpt.Text = helloContent
				default:
					if len(content) == 0 {
						msgOpt.Text = "您没有输入有效内容哟。 :wink: \n" + helpContent
					} else {
						imgURL := EncodeURL(content)
						logger.Info("generated", zap.String("url", imgURL))
						msgOpt.Text = ":hugging_face: 这是您的:horse: ："
						msgOpt.Attachments = []openapi.MessageAttachment{
							{
								Images: []openapi.MessageAttachmentImage{
									{
										Url: &imgURL,
									},
								},
							},
						}
					}
				}
				outgoing, _, badLuck := botAPI.Message.Create(ctx, &msgOpt)
				cancel()
				if badLuck != nil {
					logger.Error("failed to make response", zap.Any("outgoing msg", outgoing), zap.Error(badLuck))
					continue
				}
			}
		}
	}
}

// EncodeURL encoded
func EncodeURL(content string) (URL string) {
	var values = url.Values{
		"content": []string{content},
		"size":    []string{strconv.FormatInt(int64(C.QRCodeSize), 10)},
		"src":     []string{srctag},
	}
	URL = C.EncodeAPIEndpoint + values.Encode()
	return
}

// DownloadImageAndScan from imageURL and scan it
func DownloadImageAndScan(imageClue string) (result []string, err error) {
	sepIdx := strings.LastIndex(imageClue, "/")
	if sepIdx == -1 {
		err = errors.New("imageURL malformed")
		return
	}
	values := url.Values{
		"file_key": []string{imageClue[sepIdx+1:]},
		"token":    []string{C.RTMToken},
	}
	req, err := http.NewRequest(http.MethodGet, fileAPIEndpoint+values.Encode(), http.NoBody)
	if err != nil {
		err = errors.Wrap(err, "http.NewRequest")
		return
	}
	req.Header.Set("User-Agent", "QR Code Bot(XiaoMa)"+Version)
	resp, err := downloader.Do(req)
	if err != nil {
		err = errors.Wrap(err, "downloader.Do")
		return
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	_, copyErr := io.CopyN(&buf, resp.Body, int64(C.MaxDecodeFileSize))
	if copyErr == nil {
		err = errors.New("body is bigger than MaxDecodeFileSize")
		return
	}
	if copyErr != io.EOF {
		err = errors.Wrap(copyErr, "io.CopyN")
		return
	}

	img, _, err := image.Decode(&buf)
	if err != nil {
		err = errors.Wrap(err, "image.Decode")
		return
	}

	result, err = qrcode.DecodeQRCode(img)
	return
}
