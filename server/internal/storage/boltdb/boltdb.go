// Copyright 2022 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package boltdb

import (
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/ristretto/z"
	"github.com/tidwall/match"
	"go.chensl.me/redix/server/internal/storage"
	"go.chensl.me/redix/server/pkg/bytesconv"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
)

var (
	_defaultBucket = []byte("default")
)

type boltDBStorage struct {
	db        *bbolt.DB
	expiresCh chan []byte
	closer    *z.Closer
	logger    *zap.Logger
}

func NewStorage(path string, logger *zap.Logger) (storage.Interface, error) {
	dir := filepath.Dir(path)
	if !exist(dir) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	db, err := bbolt.Open(path, 0600, &bbolt.Options{
		NoFreelistSync: true,
		FreelistType:   bbolt.FreelistMapType,
	})
	if err != nil {
		return nil, err
	}
	s := &boltDBStorage{
		db:        db,
		expiresCh: make(chan []byte, 1),
		closer:    z.NewCloser(1),
		logger:    logger,
	}
	go s.asyncDeleter()
	return s, nil
}

func (s *boltDBStorage) Keys(pattern string) ([][]byte, error) {
	var keys [][]byte

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(_defaultBucket)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			if match.Match(bytesconv.BytesToString(k), pattern) {
				keys = append(keys, cloneBytes(k))
			}
			return nil
		})
	})

	return keys, err
}

func (s *boltDBStorage) Del(keys ...[]byte) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	var cnt int
	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(_defaultBucket)
		if b == nil {
			return nil
		}

		for _, k := range keys {
			if b.Get(k) == nil {
				continue
			}
			cnt++
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})

	return cnt, err
}

func (s *boltDBStorage) DropAll() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket(_defaultBucket); err != nil && err != bbolt.ErrBucketNotFound {
			return err
		}
		return nil
	})
}

func (s *boltDBStorage) Expire(key []byte, dur time.Duration) error {
	if dur <= 0 {
		return storage.ErrInvalidOpts
	}

	err := s.db.Update(func(tx *bbolt.Tx) error {
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

		ent.ExpiresAt = time.Now().Add(dur).Unix()
		return s.putEntry(b, key, ent)
	})

	return err
}

func (s *boltDBStorage) TTL(key []byte) (int64, error) {
	ttl := int64(-1)

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(_defaultBucket)
		if b == nil {
			ttl = -2
			return nil
		}

		ent, err := s.getEntry(b, key)
		if err != nil {
			return err
		}
		if ent == nil {
			ttl = -2
			return nil
		}

		if ent.ExpiresAt > 0 {
			ttl = int64(time.Until(time.Unix(ent.ExpiresAt, 0)).Seconds())
		}
		return nil
	})

	return ttl, err
}

func (s *boltDBStorage) Close() error {
	s.logger.Info("stopping asyncDeleter")
	s.closer.SignalAndWait()
	return s.db.Close()
}

func (s *boltDBStorage) asyncDeleter() {
	defer s.closer.Done()
	for {
		select {
		case <-s.closer.HasBeenClosed():
			return
		case key := <-s.expiresCh:
			err := s.db.Update(func(tx *bbolt.Tx) error {
				b := tx.Bucket(_defaultBucket)
				if b == nil {
					return nil
				}
				return b.Delete(key)
			})
			if err != nil {
				s.logger.Error("failed to expires key",
					zap.ByteString("key", key),
					zap.Error(err),
				)
			}
		}
	}
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func exist(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
