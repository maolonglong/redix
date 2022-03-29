// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package badger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.chensl.me/redix/server/internal/storage"
)

func Test_badgerStorage_CommonCmd(t *testing.T) {
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

	ttl, err := s.TTL([]byte("key"))
	assert.Equal(t, int64(-2), ttl)
	assert.NoError(t, err)

	err = s.Expire([]byte("key"), -1)
	assert.ErrorIs(t, err, storage.ErrInvalidOpts)
}
