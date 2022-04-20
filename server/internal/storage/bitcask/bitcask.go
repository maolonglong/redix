// Copyright 2022 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package bitcask

import (
	"time"

	"github.com/dgraph-io/ristretto/z"
	"github.com/tidwall/match"
	"go.chensl.me/bitcask"
	"go.chensl.me/redix/server/internal/storage"
	"go.chensl.me/redix/server/pkg/bytesconv"
	"go.uber.org/zap"
)

type bitcaskStorage struct {
	db     *bitcask.DB
	closer *z.Closer
	logger *zap.Logger
}

func NewStorage(path string, logger *zap.Logger) (storage.Interface, error) {
	db, err := bitcask.Open(path)
	if err != nil {
		return nil, err
	}
	s := &bitcaskStorage{
		db:     db,
		closer: z.NewCloser(1),
		logger: logger,
	}
	go s.gc()
	return s, nil
}

func (s *bitcaskStorage) Keys(pattern string) ([][]byte, error) {
	var keys [][]byte

	err := s.db.ForEach(func(key, _ []byte) error {
		if match.Match(bytesconv.BytesToString(key), pattern) {
			keys = append(keys, key)
		}
		return nil
	})

	return keys, err
}

func (s *bitcaskStorage) Del(keys ...[]byte) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	var cnt int
	for _, k := range keys {
		err := s.db.Delete(k)
		if err == bitcask.ErrNotExist {
			continue
		}
		if err != nil {
			return cnt, err
		}
		cnt++
	}

	return cnt, nil
}

func (s *bitcaskStorage) DropAll() error {
	// TODO: 直接删文件，在 bitcask 库里实现
	panic("unimplemented")
}

func (s *bitcaskStorage) Expire(key []byte, dur time.Duration) error {
	if dur <= 0 {
		return storage.ErrInvalidOpts
	}

	entry, err := s.getEntry(key)
	if err != nil {
		return err
	}

	entry.ExpiresAt = time.Now().Add(dur).Unix()

	return s.putEntry(key, entry)
}

func (s *bitcaskStorage) TTL(key []byte) (int64, error) {
	entry, err := s.getEntry(key)
	if err == storage.ErrNotExist {
		return -2, nil
	}
	if err != nil {
		return 0, err
	}

	if entry.ExpiresAt > 0 {
		return int64(time.Until(time.Unix(entry.ExpiresAt, 0)).Seconds()), nil
	}

	return -1, nil
}

func (s *bitcaskStorage) Close() error {
	s.logger.Info("graceful shutdown...")
	s.closer.SignalAndWait()
	return s.db.Close()
}

func (s *bitcaskStorage) gc() {
	defer s.closer.Done()
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-s.closer.HasBeenClosed():
			ticker.Stop()
			return
		case <-ticker.C:
		}
		if err := s.db.Reclaim(); err != nil {
			s.logger.Error("failed to reclaim", zap.Error(err))
		}
	}
}
