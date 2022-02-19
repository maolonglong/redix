// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package redix

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

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

func Default() (*Server, error) {
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

	addr := fmt.Sprintf("tcp://%s:%d", viper.GetString("host"), viper.GetInt("port"))

	// TODO: graceful shutdown evio server
	//nolint:errcheck
	go evio.Serve(events, addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	s.logger.Info("redix server listening", zap.String("addr", addr))
	sig := <-quit
	s.logger.Info("graceful shutdown", zap.Stringer("signal", sig))

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
					out = redcon.AppendError(out, "ERR unknown command '"+string(args[0])+"'")
				}
			case "AUTH":
				if len(args) != 2 {
					out = redcon.AppendError(out, "ERR wrong number of arguments for '"+string(args[0])+"' command")
				} else if bytesconv.BytesToString(args[1]) != s.password {
					out = redcon.AppendError(out, "ERROR WRONGPASS invalid username-password pair or user is disabled.")
				} else {
					c.auth = true
					out = redcon.AppendOK(out)
				}
			case "PING":
				if len(args) > 2 {
					out = redcon.AppendError(out, "ERR wrong number of arguments for '"+string(args[0])+"' command")
				} else if len(args) == 2 {
					out = redcon.AppendBulk(out, args[1])
				} else {
					out = redcon.AppendString(out, "PONG")
				}
			case "ECHO":
				if len(args) != 2 {
					out = redcon.AppendError(out, "ERR wrong number of arguments for '"+string(args[0])+"' command")
				} else {
					out = redcon.AppendBulk(out, args[1])
				}
			case "QUIT":
				out = redcon.AppendString(out, "OK")
				action = evio.Close
			}
		}
	}
	c.is.End(data)
	return //nolint:nakedret
}
