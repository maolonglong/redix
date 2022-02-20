// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package redix

import (
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/match"
	"go.chensl.me/redix/internal/bytesconv"
	"go.chensl.me/redix/pkg/storage"
	"go.uber.org/zap"
)

func (s *Server) initCommands() {
	s.register("keys", s.cmdKEYS)
	s.register("ttl", s.cmdTTL)
	s.register("expire", s.cmdEXPIRE)
	s.register("del", s.cmdDEL)
	s.register("flushall", s.cmdFLUSHALL)
	s.register("flushdb", s.cmdFLUSHALL)

	s.register("set", s.cmdSET)
	s.register("setnx", s.cmdSETNX)
	s.register("get", s.cmdGET)
	s.register("incr", s.cmdAdd(1))
	s.register("decr", s.cmdAdd(-1))
	s.register("incrby", s.cmdAddBy(true))
	s.register("decrby", s.cmdAddBy(false))
}

func (s *Server) cmdSET(c *Context) {
	if len(c.Args) < 2 {
		c.ErrInvalidArgs()
		return
	}

	var (
		key    = c.Args[0]
		val    = c.Args[1]
		nx     bool
		xx     bool
		ttlSet bool
		ttl    time.Duration
	)
	args := c.Args[2:]
	for len(args) > 0 {
		timeUnit := time.Second
		switch strings.ToUpper(bytesconv.BytesToString(args[0])) {
		case "PX":
			timeUnit = time.Microsecond
			fallthrough
		case "EX":
			if len(args) < 2 {
				c.ErrSyntax()
				return
			}
			if ttlSet {
				c.ErrSyntax()
				return
			}
			exp, err := strconv.Atoi(bytesconv.BytesToString(args[1]))
			if err != nil {
				c.ErrInvalidInt()
				return
			}
			if exp <= 0 {
				c.ErrInvalidExp()
				return
			}
			ttl = time.Duration(exp) * timeUnit
			ttlSet = true
			args = args[2:]
		case "NX":
			nx = true
			args = args[1:]
		case "XX":
			xx = true
			args = args[1:]
		default:
			c.ErrSyntax()
			return
		}
	}

	if nx && xx {
		c.ErrSyntax()
		return
	}

	err := s.store.Set(key, val, storage.SetOptions{
		TTL: ttl,
		NX:  nx,
		XX:  xx,
	})
	if err == storage.ErrExist || err == storage.ErrNotExist {
		c.AppendNull()
		return
	}
	if err != nil {
		s.logUnknownError("store.Set", err)
		c.ErrUnknown(err)
		return
	}

	c.AppendOK()
}

func (s *Server) cmdGET(c *Context) {
	if len(c.Args) != 1 {
		c.ErrInvalidArgs()
		return
	}

	v, err := s.store.Get(c.Args[0])
	if err == storage.ErrNotExist {
		c.AppendNull()
		return
	}
	if err != nil {
		s.logUnknownError("store.Get", err)
		c.ErrUnknown(err)
		return
	}

	c.AppendBulk(v)
}

func (s *Server) cmdTTL(c *Context) {
	if len(c.Args) != 1 {
		c.ErrInvalidArgs()
		return
	}

	ttl, err := s.store.TTL(c.Args[0])
	if err != nil {
		s.logUnknownError("store.TTL", err)
		c.ErrUnknown(err)
		return
	}

	c.AppendInt(ttl)
}

func (s *Server) cmdEXPIRE(c *Context) {
	if len(c.Args) != 2 {
		c.ErrInvalidArgs()
		return
	}

	exp, err := strconv.Atoi(bytesconv.BytesToString(c.Args[1]))
	if err != nil {
		c.ErrInvalidInt()
		return
	}
	if exp <= 0 {
		c.ErrInvalidExp()
		return
	}

	err = s.store.Expire(c.Args[0], time.Duration(exp)*time.Second)
	if err == storage.ErrNotExist {
		c.AppendInt(0)
		return
	}
	if err != nil {
		s.logUnknownError("store.Expire", err)
		c.ErrUnknown(err)
		return
	}

	c.AppendInt(1)
}

func (s *Server) cmdDEL(c *Context) {
	if len(c.Args) == 0 {
		c.ErrInvalidArgs()
		return
	}

	n, err := s.store.Del(c.Args...)
	if err != nil {
		s.logUnknownError("store.Del", err)
		c.ErrUnknown(err)
		return
	}

	c.AppendInt(int64(n))
}

func (s *Server) cmdFLUSHALL(c *Context) {
	if len(c.Args) != 0 {
		c.ErrInvalidArgs()
		return
	}

	err := s.store.DropAll()
	if err != nil {
		s.logUnknownError("store.FlushAll", err)
		c.ErrUnknown(err)
		return
	}

	c.AppendOK()
}

func (s *Server) cmdKEYS(c *Context) {
	if len(c.Args) != 1 {
		c.ErrInvalidArgs()
		return
	}

	pat := bytesconv.BytesToString(c.Args[0])
	if !match.IsPattern(pat) {
		c.ErrSyntax()
		return
	}

	keys, err := s.store.Keys(pat)
	if err != nil {
		s.logUnknownError("store.Keys", err)
		c.ErrUnknown(err)
		return
	}

	c.AppendBulkArray(keys)
}

func (s *Server) cmdSETNX(c *Context) {
	if len(c.Args) != 2 {
		c.ErrInvalidArgs()
		return
	}

	err := s.store.Set(c.Args[0], c.Args[1], storage.SetOptions{NX: true})
	if err == storage.ErrExist {
		c.AppendInt(0)
		return
	}
	if err != nil {
		s.logUnknownError("store.Set", err)
		c.ErrUnknown(err)
		return
	}

	c.AppendInt(1)
}

func (s *Server) cmdAdd(delta int) CommandFunc {
	return func(c *Context) {
		if len(c.Args) != 1 {
			c.ErrInvalidArgs()
			return
		}

		i, err := s.store.Add(c.Args[0], delta)
		if err == storage.ErrInvalidInt {
			c.ErrInvalidInt()
			return
		}
		if err != nil {
			s.logUnknownError("store.Add", err)
			c.ErrUnknown(err)
			return
		}

		c.AppendInt(int64(i))

	}
}

func (s *Server) cmdAddBy(incr bool) CommandFunc {
	return func(c *Context) {
		if len(c.Args) != 2 {
			c.ErrInvalidArgs()
			return
		}

		delta, err := strconv.Atoi(bytesconv.BytesToString(c.Args[1]))
		if err != nil {
			c.ErrInvalidInt()
			return
		}

		if !incr {
			delta = -delta
		}

		i, err := s.store.Add(c.Args[0], delta)
		if err == storage.ErrInvalidInt {
			c.ErrInvalidInt()
			return
		}
		if err != nil {
			s.logUnknownError("store.Add", err)
			c.ErrUnknown(err)
			return
		}

		c.AppendInt(int64(i))
	}
}

func (s *Server) logUnknownError(method string, err error) {
	s.logger.Error("unknown error",
		zap.String("method", method),
		zap.Error(err),
	)
}
