/*
 * Copyright (c) 2018 LI Zhennan
 *
 * Use of this work is governed by an MIT License.
 * You may find a license copy in project root.
 */

package main

import (
	"errors"
)

// lack is a dummy error provider
func lack(field string) error {
	return errors.New("lack of field " + field)
}
