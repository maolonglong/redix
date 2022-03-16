// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package storage

import "time"

type SetOptions struct {
	TTL time.Duration
	NX  bool
	XX  bool
}
