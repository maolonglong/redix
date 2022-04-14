// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package badger

import (
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v3"
	"go.chensl.me/redix/server/internal/storage"
	"go.chensl.me/redix/server/pkg/bytesconv"
)

func (s *badgerStorage) Set(key, value []byte, opts storage.SetOptions) error {
	if opts.NX && opts.XX {
		return storage.ErrInvalidOpts
	}

	err := s.db.Update(func(txn *badger.Txn) error {
		if opts.NX || opts.XX {
			_, err := txn.Get(key)
			if err == badger.ErrKeyNotFound {
				if opts.XX {
					return storage.ErrNotExist
				}
			} else if err == nil {
				if opts.NX {
					return storage.ErrExist
				}
			} else {
				return err
			}
		}

		e := badger.NewEntry(key, value)
		if opts.TTL > 0 {
			e = e.WithTTL(opts.TTL)
		}

		return txn.SetEntry(e)
	})

	return err
}

func (s *badgerStorage) Get(key []byte) ([]byte, error) {
	var val []byte

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return storage.ErrNotExist
		}
		if err != nil {
			return err
		}
		val, _ = item.ValueCopy(nil)
		return nil
	})

	return val, err
}

func (s *badgerStorage) Add(key []byte, delta int) (int, error) {
	var i int

	err := s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			i = delta
			return txn.Set(key, bytesconv.StringToBytes(strconv.Itoa(delta)))
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			i, err = strconv.Atoi(bytesconv.BytesToString(val))
			if err != nil {
				return storage.ErrInvalidInt
			}
			i += delta
			e := badger.NewEntry(key, bytesconv.StringToBytes(strconv.Itoa(i)))
			if exp := item.ExpiresAt(); exp > 0 {
				e = e.WithTTL(time.Until(time.Unix(int64(exp), 0)))
			}
			return txn.SetEntry(e)
		})
	})

	return i, err
}
