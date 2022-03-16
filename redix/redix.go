// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package redix

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/tidwall/evio"
	"github.com/tidwall/redcon"
	"go.chensl.me/redix/internal/bytesconv"
	"go.chensl.me/redix/pkg/storage"
	"go.chensl.me/redix/pkg/storage/badger"
	"go.uber.org/zap"
)

type CommandFunc func(c *Context)

type Server struct {
	commands map[string]CommandFunc
	password string
	store    storage.Interface
	logger   *zap.Logger
}

func New() (*Server, error) {
	srv := &Server{
		commands: make(map[string]CommandFunc),
		password: viper.GetString("password"),
	}
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	srv.logger = logger
	srv.store, err = badger.NewStorage(viper.GetString("data_dir"), logger)
	if err != nil {
		return nil, err
	}
	srv.initCommands()
	return srv, nil
}

func (s *Server) Run() error {
	events := evio.Events{
		NumLoops: 1,
		Opened:   s.openedHandler,
		Data:     s.dataHandler,
	}

	path, err := filepath.Abs(viper.GetString("data_dir"))
	if err != nil {
		s.logger.Error("failed to get absolute path", zap.Error(err))
		return err
	}

	s.logger.Info("redix server started",
		zap.String("host", viper.GetString("host")),
		zap.Int("port", viper.GetInt("port")),
		zap.String("data_dir", path),
	)

	addr := fmt.Sprintf("tcp://%s:%d", viper.GetString("host"), viper.GetInt("port"))
	return evio.Serve(events, addr)
}

func (s *Server) Cleanup() error {
	return s.store.Close()
}

func (s *Server) register(cmd string, fn CommandFunc) {
	s.commands[strings.ToUpper(cmd)] = fn
}

func (s *Server) openedHandler(ec evio.Conn) (out []byte, opts evio.Options, action evio.Action) {
	ec.SetContext(new(Context))
	return //nolint:nakedret
}

func (s *Server) dataHandler(ec evio.Conn, in []byte) (out []byte, action evio.Action) {
	defer func() {
		if err := recover(); err != nil {
			s.logger.Error("panic",
				zap.ByteString("in", in),
				zap.Binary("raw", in),
				zap.Any("err", err),
			)
			out = redcon.AppendError(out, "ERR panic.")
			action = evio.Close
		}
	}()

	c := ec.Context().(*Context)
	data := c.is.Begin(in)
	var complete bool
	var err error
	var args [][]byte
	for action == evio.None {
		complete, args, _, data, err = redcon.ReadNextCommand(data, args[:0])
		if err != nil {
			action = evio.Close
			out = redcon.AppendError(out, err.Error())
			break
		}
		if !complete {
			break
		}
		cmd := strings.ToUpper(bytesconv.BytesToString(args[0]))
		if cmd != "AUTH" && s.password != "" && !c.auth {
			out = redcon.AppendError(out, "ERROR Authentication required.")
			continue
		}
		if len(args) > 0 {
			switch cmd {
			default:
				if fn, ok := s.commands[cmd]; ok {
					c.cmd = args[0]
					c.Args = args[1:]
					c.out = &out
					fn(c)
				} else {
					s.logger.Warn("unknown command",
						zap.ByteString("cmd", args[0]),
						zap.ByteStrings("args", args[1:]),
					)
					out = redcon.AppendError(out, "ERR unknown command '"+string(args[0])+"'.")
				}
			case "AUTH":
				if len(args) != 2 {
					out = redcon.AppendError(out, "ERR wrong number of arguments for '"+string(args[0])+"' command.")
				} else if bytesconv.BytesToString(args[1]) != s.password {
					out = redcon.AppendError(out, "ERROR WRONGPASS invalid username-password pair or user is disabled.")
				} else {
					c.auth = true
					out = redcon.AppendOK(out)
				}
			case "PING":
				if len(args) > 2 {
					out = redcon.AppendError(out, "ERR wrong number of arguments for '"+string(args[0])+"' command.")
				} else if len(args) == 2 {
					out = redcon.AppendBulk(out, args[1])
				} else {
					out = redcon.AppendString(out, "PONG")
				}
			case "QUIT":
				out = redcon.AppendOK(out)
				action = evio.Close
			case "SHUTDOWN":
				out = redcon.AppendOK(out)
				action = evio.Shutdown
			}
		}
	}
	c.is.End(data)
	return //nolint:nakedret
}
