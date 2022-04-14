// Copyright 2022 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package boltdb

import (
	"strconv"
	"time"

	"go.chensl.me/redix/server/internal/storage"
	"go.chensl.me/redix/server/internal/storage/boltdb/entrypb"
	"go.chensl.me/redix/server/pkg/bytesconv"
	"go.etcd.io/bbolt"
)

func (s *boltDBStorage) Set(key, value []byte, opts storage.SetOptions) error {
	if opts.NX && opts.XX {
		return storage.ErrInvalidOpts
	}

	err := s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(_defaultBucket)
		if err != nil {
			return err
		}

		if opts.NX || opts.XX {
			ent, err := s.getEntry(b, key)
			if err != nil {
				return err
			}

			if ent == nil {
				if opts.XX {
					return storage.ErrNotExist
				}
			} else {
				if opts.NX {
					return storage.ErrExist
				}
			}
		}

		ent := &entrypb.Entry{Value: value}
		if opts.TTL > 0 {
			ent.ExpiresAt = time.Now().Add(opts.TTL).Unix()
		}

		return s.putEntry(b, key, ent)
	})

	return err
}

func (s *boltDBStorage) Get(key []byte) ([]byte, error) {
	var val []byte

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(_defaultBucket)
		if b == nil {
			return storage.ErrNotExist
		}

		ent, err := s.getEntry(b, key)
		if err != nil {
			return err
		}
		if ent == nil {
			return storage.ErrNotExist
		}

		val = ent.Value
		return nil
	})

	return val, err
}

func (s *boltDBStorage) Add(key []byte, delta int) (int, error) {
	var i int

	err := s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(_defaultBucket)
		if err != nil {
			return err
		}

		ent, err := s.getEntry(b, key)
		if err != nil {
			return err
		}

		if ent == nil {
			i = delta
			return s.putEntry(b, key, &entrypb.Entry{Value: bytesconv.StringToBytes(strconv.Itoa(delta))})
		}

		i, err = strconv.Atoi(bytesconv.BytesToString(ent.Value))
		if err != nil {
			return storage.ErrInvalidInt
		}
		i += delta
		ent.Value = bytesconv.StringToBytes(strconv.Itoa(i))
		return s.putEntry(b, key, ent)
	})

	return i, err
}
