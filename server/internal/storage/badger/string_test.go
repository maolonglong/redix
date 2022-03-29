// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package badger

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.chensl.me/redix/server/internal/storage"
)

func Test_badgerStorage_StringCmd(t *testing.T) {
	path, err := os.MkdirTemp(os.TempDir(), "*")
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewStorage(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		s.Close()
		os.RemoveAll(path)
	}()

	key, value := []byte("key"), []byte("value")

	v, err := s.Get(key)
	assert.Nil(t, v)
	assert.ErrorIs(t, err, storage.ErrNotExist)

	err = s.Expire(key, time.Minute)
	assert.ErrorIs(t, err, storage.ErrNotExist)

	err = s.Set(key, value, storage.SetOptions{
		NX: true,
		XX: true,
	})
	assert.ErrorIs(t, err, storage.ErrInvalidOpts)

	err = s.Set(key, value, storage.SetOptions{
		XX: true,
	})
	assert.ErrorIs(t, err, storage.ErrNotExist)

	err = s.Set(key, value, storage.SetOptions{})
	assert.NoError(t, err)

	ttl, err := s.TTL(key)
	assert.Equal(t, int64(-1), ttl)
	assert.NoError(t, err)

	err = s.Expire(key, time.Minute)
	assert.NoError(t, err)

	ttl, err = s.TTL(key)
	assert.Equal(t, int64(59), ttl)
	assert.NoError(t, err)

	v, err = s.Get(key)
	assert.Equal(t, value, v)
	assert.NoError(t, err)

	err = s.Set(key, value, storage.SetOptions{
		NX: true,
	})
	assert.ErrorIs(t, err, storage.ErrExist)

	err = s.Set(key, value, storage.SetOptions{
		TTL: 2 * time.Second,
	})
	assert.NoError(t, err)

	v, err = s.Get(key)
	assert.Equal(t, value, v)
	assert.NoError(t, err)

	ttl, err = s.TTL(key)
	assert.Equal(t, int64(1), ttl)
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)
	v, err = s.Get(key)
	assert.Nil(t, v)
	assert.ErrorIs(t, err, storage.ErrNotExist)
}
