/*
 * Copyright (c) 2018 LI Zhennan
 *
 * Use of this work is governed by an MIT License.
 * You may find a license copy in project root.
 */

package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/pkg/errors"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/nanmu42/qrcode-api"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// requestCounter is used to count request num
// should only be affected by atomic action
var requestCounter uint64

func setupRouter() (router *gin.Engine) {
	if C.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router = gin.New()
	router.HandleMethodNotAllowed = true
	router.Use(gin.Recovery())
	// set debug mode if in need
	if C.Debug {
		router.Use(gin.Logger())
	}
	// log requests
	router.Use(RequestLogger(logger))

	// setup routes
	router.GET("/encode", EncodeQRCode)
	router.POST("/decode", DecodeQRCode)
	return
}

func startAPI(handler http.Handler, port string) {
	// timeout for safe exit
	const shutdownTimeout = 2 * time.Minute

	var exitSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}

	server := &http.Server{
		Addr:    port,
		Handler: handler,
		// ReadTimeout time limit to read the request
		ReadTimeout: 10 * time.Second,
		// WriteTimeout time limit for request reading and response
		WriteTimeout: 40 * time.Second,
		// IdleTimeout keep-alive waiting time
		IdleTimeout: 60 * time.Second,
		// MaxHeaderBytes max header is 8KB
		MaxHeaderBytes: 1 << 13,
	}

	go func() {
		// service connections
		fmt.Println("API starting...")
		logger.Info("API starting...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("API HTTP service: %v", err)
			logger.Fatal("API HTTP service fatal error",
				zap.Error(err),
			)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of `shutdownTimeout` seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, exitSignals...)
	<-quit
	fmt.Println("API is exiting safely...")
	logger.Info("API is exiting safely...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("API exiting timed out:", err)
		logger.Fatal("API exiting timed out",
			zap.Error(err),
		)
	}
	logger.Info("API exited successfully. :)")
	fmt.Println("API exited successfully. :)")

	return
}

// RequestLogger logs every request via zap
func RequestLogger(l *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// request timer
		receivedAt := time.Now()
		// counter++
		sequence := atomic.AddUint64(&requestCounter, 1)

		// before
		c.Next()
		// after
		l.Info("request",
			zap.Uint64("seq", sequence),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("IP", c.ClientIP()),
			zap.String("UA", c.Request.UserAgent()),
			zap.String("ref", c.Request.Referer()),
			zap.Duration("lapse", time.Now().Sub(receivedAt)),
			zap.Strings("err", c.Errors.Errors()),
		)
	}
}

// EncodeQRCode controller to encode QR code per request
func EncodeQRCode(c *gin.Context) {
	var err error

	encoder, err := ParseEncodeRequest(c.Request.URL.Query())
	if err != nil {
		c.Error(err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	var buf bytes.Buffer
	gotType, err := encoder.Encode(&buf)
	if err != nil {
		c.Error(err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	switch gotType {
	case qrcode.TypePNG:
		c.DataFromReader(http.StatusOK, int64(buf.Len()), "image/png", &buf, map[string]string{})
	case qrcode.TypeString:
		c.DataFromReader(http.StatusOK, int64(buf.Len()), "text/plain; charset=utf-8", &buf, map[string]string{})
	}

	return
}

// DecodeQRCode controller to decode QR Code
func DecodeQRCode(c *gin.Context) {
	var err error

	// avoid too big image
	if c.Request.ContentLength >= maxDecodeFileByte {
		err = errors.New("request is too big(content-length)")
		c.Error(err)
		c.Request.Body.Close()
		c.JSON(http.StatusRequestEntityTooLarge, DecodeResponse{
			OK:      false,
			Desc:    "request is too big",
			Content: nil,
		})
		return
	}
	var buf bytes.Buffer
	_, readErr := io.CopyN(&buf, c.Request.Body, maxDecodeFileByte)
	if readErr == nil {
		err = errors.New("request is too big(actually read)")
		c.Error(err)
		c.Request.Body.Close()
		c.JSON(http.StatusRequestEntityTooLarge, DecodeResponse{
			OK:      false,
			Desc:    "request is too big",
			Content: nil,
		})
		return
	}
	if readErr != io.EOF {
		err = errors.Wrap(readErr, "body read error")
		c.Error(err)
		c.JSON(http.StatusOK, DecodeResponse{
			OK:      false,
			Desc:    err.Error(),
			Content: nil,
		})
		return
	}

	// decode image
	input, _, err := image.Decode(&buf)
	if err != nil {
		err = errors.Wrap(err, "file decoding error")
		c.Error(err)
		c.JSON(http.StatusOK, DecodeResponse{
			OK:      false,
			Desc:    err.Error(),
			Content: nil,
		})
		return
	}

	contents, err := qrcode.DecodeQRCode(input)
	if err != nil {
		err = errors.Wrap(err, "QR Code scanning error")
		c.Error(err)
		c.JSON(http.StatusOK, DecodeResponse{
			OK:      false,
			Desc:    err.Error(),
			Content: nil,
		})
		return
	}

	c.JSON(http.StatusOK, DecodeResponse{
		OK:      true,
		Desc:    "",
		Content: contents,
	})

	return
}
