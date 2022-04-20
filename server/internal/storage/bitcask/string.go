// Copyright 2022 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package bitcask

import (
	"strconv"
	"time"

	"go.chensl.me/redix/server/internal/storage"
	"go.chensl.me/redix/server/internal/storage/entrypb"
	"go.chensl.me/redix/server/pkg/bytesconv"
)

func (s *bitcaskStorage) Set(key, value []byte, opts storage.SetOptions) error {
	if opts.NX && opts.XX {
		return storage.ErrInvalidOpts
	}

	entry, err := s.getEntry(key)
	if err != nil && err != storage.ErrNotExist {
		return err
	}

	if err == storage.ErrNotExist {
		if opts.XX {
			return storage.ErrNotExist
		}
	} else {
		if opts.NX {
			return storage.ErrExist
		}
	}

	entry = &entrypb.Entry{Value: value}
	if opts.TTL > 0 {
		entry.ExpiresAt = time.Now().Add(opts.TTL).Unix()
	}

	return s.putEntry(key, entry)
}

func (s *bitcaskStorage) Get(key []byte) ([]byte, error) {
	entry, err := s.getEntry(key)
	if err != nil {
		return nil, err
	}

	return entry.Value, nil
}

func (s *bitcaskStorage) Add(key []byte, delta int) (int, error) {
	entry, err := s.getEntry(key)
	if err == storage.ErrNotExist {
		return delta, s.putEntry(key, &entrypb.Entry{Value: bytesconv.StringToBytes(strconv.Itoa(delta))})
	}
	if err != nil {
		return 0, err
	}

	i, err := strconv.Atoi(bytesconv.BytesToString(entry.Value))
	if err != nil {
		return 0, storage.ErrInvalidInt
	}

	i += delta
	entry.Value = bytesconv.StringToBytes(strconv.Itoa(i))
	return i, s.putEntry(key, entry)
}
