// Copyright 2022 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package bitcask

import (
	"time"

	"go.chensl.me/bitcask"
	"go.chensl.me/redix/server/internal/storage"
	"go.chensl.me/redix/server/internal/storage/entrypb"
	"google.golang.org/protobuf/proto"
)

func (s *bitcaskStorage) putEntry(key []byte, entry *entrypb.Entry) error {
	b, err := proto.Marshal(entry)
	if err != nil {
		return err
	}
	return s.db.Put(key, b)
}

func (s *bitcaskStorage) getEntry(key []byte) (*entrypb.Entry, error) {
	b, err := s.db.Get(key)
	if err == bitcask.ErrNotExist {
		return nil, storage.ErrNotExist
	}
	if err != nil {
		return nil, err
	}
	var entry entrypb.Entry
	if err := proto.Unmarshal(b, &entry); err != nil {
		return nil, err
	}
	if entry.ExpiresAt > 0 && time.Now().Unix() >= entry.ExpiresAt {
		_ = s.db.Delete(key)
		return nil, storage.ErrNotExist
	}
	return &entry, nil
}
