// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"

	"github.com/tidwall/evio"
	"github.com/tidwall/redcon"
)

type Context struct {
	is   evio.InputStream
	out  *[]byte
	cmd  []byte
	auth bool
	Args [][]byte
}

func (c *Context) AppendError(s string) {
	*c.out = redcon.AppendError(*c.out, s)
}

func (c *Context) AppendBulkArray(bs [][]byte) {
	*c.out = redcon.AppendArray(*c.out, len(bs))
	for _, b := range bs {
		*c.out = redcon.AppendBulk(*c.out, b)
	}
}

func (c *Context) AppendInt(i int64) {
	*c.out = redcon.AppendInt(*c.out, i)
}

func (c *Context) AppendOK() {
	*c.out = redcon.AppendOK(*c.out)
}

func (c *Context) AppendNull() {
	*c.out = redcon.AppendNull(*c.out)
}

func (c *Context) AppendBulk(b []byte) {
	*c.out = redcon.AppendBulk(*c.out, b)
}

func (c *Context) ErrSyntax() {
	c.AppendError("ERR syntax error")
}

func (c *Context) ErrInvalidArgs() {
	c.AppendError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", c.cmd))
}

func (c *Context) ErrInvalidInt() {
	c.AppendError("ERR value is not an integer or out of range")
}

func (c *Context) ErrInvalidExp() {
	c.AppendError("ERR invalid expire time in set")
}

func (c *Context) ErrUnknown(err error) {
	c.AppendError(fmt.Sprintf("Err unknown %q", err.Error()))
}
