// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package storage

import "errors"

var (
	ErrInvalidOpts = errors.New("invalid options")
	ErrExist       = errors.New("key already exists")
	ErrNotExist    = errors.New("key does not exist")
	ErrInvalidInt  = errors.New("invalid int")
)
