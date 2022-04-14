// Copyright 2021 MaoLongLong. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func MustInit() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/redix/")
	viper.AddConfigPath("$HOME/.redix")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("redix")
	viper.AutomaticEnv()

	viper.SetDefault("host", "0.0.0.0")
	viper.SetDefault("port", 6380)
	viper.SetDefault("password", "")
	viper.SetDefault("data_dir", "./data")
	viper.SetDefault("driver", "badger")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}
}
