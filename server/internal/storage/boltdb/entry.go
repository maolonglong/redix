// Copyright 2022 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package boltdb

import (
	"time"

	"go.chensl.me/redix/server/internal/storage/boltdb/entrypb"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

func (*boltDBStorage) putEntry(bucket *bbolt.Bucket, key []byte, ent *entrypb.Entry) error {
	v, err := proto.Marshal(ent)
	if err != nil {
		return err
	}
	return bucket.Put(key, v)
}

func (s *boltDBStorage) getEntry(bucket *bbolt.Bucket, key []byte) (*entrypb.Entry, error) {
	// TODO: return not exists error

	v := bucket.Get(key)
	if v == nil {
		return nil, nil
	}
	var ent entrypb.Entry
	if err := proto.Unmarshal(v, &ent); err != nil {
		return nil, err
	}
	if ent.ExpiresAt > 0 && time.Now().Unix() >= ent.ExpiresAt {
		s.expiresCh <- key
		return nil, nil
	}
	return &ent, nil
}
