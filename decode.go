/*
 * Copyright (c) 2018 LI Zhennan
 *
 * Use of this work is governed by an MIT License.
 * You may find a license copy in project root.
 */

package qrcode

import (
	"errors"
	"image"

	"github.com/PeterCxy/gozbar"
)

// DecodeQRCode decodes QR Code content from image
//
// content contains multiple string if there are more than one QR Code
// got decoded.
// content and err are both nil when no QR Code found.
func DecodeQRCode(img image.Image) (content []string, err error) {
	input := zbar.FromImage(img)

	s := zbar.NewScanner()
	s.SetConfig(zbar.QRCODE, zbar.CFG_ENABLE, 1)
	result := s.Scan(input)

	if result < 0 {
		err = errors.New("error occurred when scanning")
		return
	}

	input.First().Each(func(item string) {
		content = append(content, item)
	})

	return
}
