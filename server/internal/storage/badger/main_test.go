// Copyright 2022 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package badger

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	opts := []goleak.Option{
		goleak.IgnoreTopFunction("github.com/golang/glog.(*loggingT).flushDaemon"),
	}
	goleak.VerifyTestMain(m, opts...)
}
