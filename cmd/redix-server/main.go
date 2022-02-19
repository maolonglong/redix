// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	_ "embed"
	"fmt"
	"log"

	"go.chensl.me/redix"
	"go.chensl.me/redix/pkg/config"
)

//go:embed banner.txt
var banner string

var (
	branch string
	commit string
)

func main() {
	fmt.Printf("%s  branch=%s commit=%s\n\n", banner, branch, commit)
	config.MustInit()
	srv, err := redix.Default()
	if err != nil {
		log.Fatal(err)
	}
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
