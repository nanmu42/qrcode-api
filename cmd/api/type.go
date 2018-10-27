package main

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/nanmu42/qrcode-api"
)

// query filed name
const (
	contentField = "content"
	typeField    = "type"
	sizeField    = "size"
)

// ParseEncodeRequest convert encoding request to struct
func ParseEncodeRequest(values url.Values) (encoder qrcode.QREncoder, err error) {
	// required param
	encoder.Content = values.Get(contentField)
	if len(encoder.Content) == 0 {
		err = errors.New("content is empty")
		return
	}
	// max capacity of QR Code is 2953 bytes
	if len(encoder.Content) > 2048 {
		err = errors.New("content should be no more than 2KB")
		return
	}

	// optional param
	size, badNum := strconv.ParseInt(values.Get(sizeField), 10, 64)
	if badNum != nil || size <= 0 || size > int64(C.MaxEncodeWidth) {
		encoder.Size = C.DefaultEncodeWidth
	} else {
		encoder.Size = int(size)
	}
	encoder.Type = values.Get(typeField)
	return
}

// DecodeResponse content holder for response
type DecodeResponse struct {
	OK      bool     `json:"ok"`
	Desc    string   `json:"desc"`
	Content []string `json:"content"`
}
