/*
 * Copyright (c) 2018 LI Zhennan
 *
 * Use of this work is governed by an MIT License.
 * You may find a license copy in project root.
 */

package qrcode

import (
	"image/jpeg"
	"io"

	"github.com/pkg/errors"

	qrc "github.com/skip2/go-qrcode"
)

// query filed name
const (
	contentField = "content"
	typeField    = "type"
	sizeField    = "size"
)

const (
	// TypePNG file as png
	TypePNG = "png"
	// TypeJPEG file as jpeg
	TypeJPEG = "jepg"
	// TypeString QRCode
	TypeString = "string"

	// DefaultType default file type
	DefaultType = TypePNG
)

// QREncodeRequest holds info for QR code encoding
type QREncodeRequest struct {
	// content to encode
	Content string
	// desired encoding type
	Type string
	// desired image size in pixel, may not be honored
	Size int
}

// Encode produces a QR code
func (q *QREncodeRequest) Encode(dest io.Writer) (gotType string, err error) {
	qrcode, err := qrc.New(q.Content, qrc.Medium)
	if err != nil {
		err = errors.Wrap(err, "cannot get a QR Code instance")
		return
	}
	switch fileTypeCheck(q.Type) {
	case TypePNG:
		gotType = TypePNG
		err = qrcode.Write(q.Size, dest)
	case TypeJPEG:
		gotType = TypeJPEG
		err = jpeg.Encode(dest, qrcode.Image(q.Size), &jpeg.Options{
			Quality: 80,
		})
	case TypeString:
		gotType = TypeString
		_, err = dest.Write([]byte(qrcode.ToString(false)))
	}
	return
}

// fileTypeCheck checks incoming types
func fileTypeCheck(want string) string {
	switch want {
	case TypePNG, TypeJPEG, TypeString:
		return want
	default:
		return DefaultType
	}
}

// lack is a dummy error provider
func lack(field string) error {
	return errors.New("lack of field " + field)
}
