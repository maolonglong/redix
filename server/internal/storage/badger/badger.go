// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package badger

import (
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/ristretto/z"
	"github.com/dustin/go-humanize"
	"github.com/tidwall/match"
	"go.chensl.me/redix/server/internal/storage"
	"go.chensl.me/redix/server/pkg/bytesconv"
	"go.uber.org/zap"
)

type badgerStorage struct {
	db     *badger.DB
	closer *z.Closer
	logger *zap.Logger
}

func NewStorage(path string, logger *zap.Logger) (storage.Interface, error) {
	db, err := badger.Open(badger.DefaultOptions(path).WithDetectConflicts(false))
	if err != nil {
		return nil, err
	}
	s := &badgerStorage{
		db:     db,
		closer: z.NewCloser(1),
	}
	if logger != nil {
		s.logger = logger
	} else {
		s.logger, _ = zap.NewDevelopment()
	}
	go s.runValueLogGC()
	return s, nil
}

func (s *badgerStorage) Keys(pattern string) ([][]byte, error) {
	var keys [][]byte

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			k := it.Item().KeyCopy(nil)
			if match.Match(bytesconv.BytesToString(k), pattern) {
				keys = append(keys, k)
			}
		}
		return nil
	})

	return keys, err
}

func (s *badgerStorage) Del(keys ...[]byte) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	var cnt int
	err := s.db.Update(func(txn *badger.Txn) error {
		for _, k := range keys {
			_, err := txn.Get(k)
			if err == badger.ErrKeyNotFound {
				continue
			}
			if err != nil {
				return err
			}
			cnt++
			if err := txn.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})

	return cnt, err
}

func (s *badgerStorage) DropAll() error {
	return s.db.DropAll()
}

func (s *badgerStorage) Expire(key []byte, dur time.Duration) error {
	if dur <= 0 {
		return storage.ErrInvalidOpts
	}

	err := s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return storage.ErrNotExist
		}
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			e := badger.NewEntry(item.Key(), val).WithTTL(dur)
			return txn.SetEntry(e)
		})
	})

	return err
}

func (s *badgerStorage) TTL(key []byte) (int64, error) {
	ttl := int64(-1)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			ttl = -2
			return nil
		}
		if err != nil {
			return err
		}
		exp := int64(item.ExpiresAt())
		if exp > 0 {
			ttl = int64(time.Until(time.Unix(exp, 0)).Seconds())
		}
		return nil
	})

	return ttl, err
}

func (s *badgerStorage) Close() error {
	s.logger.Info("stopping value log GC")
	s.closer.SignalAndWait()
	return s.db.Close()
}

func (s *badgerStorage) runValueLogGC() {
	ticker := time.NewTicker(5 * time.Minute)
	defer s.closer.Done()
	for {
		select {
		case <-s.closer.HasBeenClosed():
			ticker.Stop()
			return
		case <-ticker.C:
		}
		lsm, vlog := s.db.Size()
		s.logger.Info("running value log GC",
			zap.String("lsm", humanize.Bytes(uint64(lsm))),
			zap.String("vlog", humanize.Bytes(uint64(vlog))),
		)
	again:
		err := s.db.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}
}
