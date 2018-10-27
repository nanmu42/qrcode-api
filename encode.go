/*
 * Copyright (c) 2018 LI Zhennan
 *
 * Use of this work is governed by an MIT License.
 * You may find a license copy in project root.
 */

package qrcode

import (
	"io"

	"github.com/pkg/errors"

	qrc "github.com/skip2/go-qrcode"
)

const (
	// TypePNG file as png
	TypePNG = "png"
	// TypeString QRCode
	TypeString = "string"

	// DefaultType default file type
	DefaultType = TypePNG
)

// QREncoder holds info for QR code encoding
type QREncoder struct {
	// content to encode
	Content string
	// desired encoding type
	Type string
	// desired image size in pixel, may not be honored
	Size int
}

// Encode produces a QR code
func (q *QREncoder) Encode(dest io.Writer) (gotType string, err error) {
	qrcode, err := qrc.New(q.Content, qrc.Medium)
	if err != nil {
		err = errors.Wrap(err, "cannot get a QR Code instance")
		return
	}
	switch fileTypeCheck(q.Type) {
	case TypePNG:
		gotType = TypePNG
		err = qrcode.Write(q.Size, dest)
	case TypeString:
		gotType = TypeString
		_, err = dest.Write([]byte(qrcode.ToString(true)))
	}
	return
}

// fileTypeCheck checks incoming types
func fileTypeCheck(want string) string {
	switch want {
	case TypePNG, TypeString:
		return want
	default:
		return DefaultType
	}
}
