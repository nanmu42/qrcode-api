/*
 * Copyright (c) 2018 LI Zhennan
 *
 * Use of this work is governed by an MIT License.
 * You may find a license copy in project root.
 */

package main

import (
	"flag"
	"fmt"

	"github.com/pkg/errors"

	"github.com/nanmu42/qrcode-api/cmd/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger     *zap.Logger
	configFile = flag.String("config", "config.toml", "config.toml file location for rly")
	// Version build params
	Version string
	// BuildDate build params
	BuildDate string
)

// maxDecodeFileByte is MaxDecodeFileSize's byte version
var maxDecodeFileByte int64

func init() {
	w := common.NewBufferedLumberjack(&lumberjack.Logger{
		Filename:   "logs/qrcode-api.log",
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
			logger.Fatal("fatal error",
				zap.Error(err),
			)
		}
	}()

	flag.Parse()

	fmt.Printf(`QRCode API(%s)
built on %s

`, Version, BuildDate)

	err = C.LoadFrom(*configFile)
	if err != nil {
		err = errors.Wrap(err, "C.LoadFrom")
		return
	}

	maxDecodeFileByte = int64(C.MaxDecodeFileSize << 10)

	router := setupRouter()
	startAPI(router, C.Port)
}
