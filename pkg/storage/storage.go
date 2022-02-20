// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package storage

import (
	"time"
)

type Interface interface {
	StringCmd

	Keys(pattern string) ([][]byte, error)
	Del(keys ...[]byte) (int, error)
	Expire(key []byte, dur time.Duration) error
	TTL(key []byte) (int64, error)
	DropAll() error
	Close() error
}

type StringCmd interface {
	Set(key, value []byte, opts SetOptions) error
	Get(key []byte) ([]byte, error)
	Add(key []byte, delta int) (int, error)
}
