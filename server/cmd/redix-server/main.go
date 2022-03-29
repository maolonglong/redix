// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/cch123/gogctuner"
	"go.chensl.me/redix/server"
	"go.chensl.me/redix/server/internal/config"
	_ "go.uber.org/automaxprocs"
)

//go:embed banner.txt
var banner string

var commit string

func main() {
	if os.Getenv("ENABLE_GC_TUNER") == "1" {
		go gogctuner.NewTuner(true /* useCgroup */, 70 /* percent */)
	}

	fmt.Printf("%s  commit=%s\n\n", banner, commit)
	config.MustInit()
	srv, err := server.New()
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Cleanup() //nolint:errcheck
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
