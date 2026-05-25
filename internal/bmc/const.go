// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"time"

	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("bmc.delay", 1)
	viper.SetDefault("bmc.fanout", 5)

	viper.SetDefault("bmc.gofish.max_concurrent_requests", 1)
	viper.SetDefault("bmc.gofish.reuse_connections", false)
	viper.SetDefault("bmc.gofish.client_timeout", time.Second*60)
	viper.SetDefault("bmc.gofish.dial_timeout", time.Second*5)
	viper.SetDefault("bmc.gofish.dial_keep_alive_timeout", time.Second*30)
	viper.SetDefault("bmc.gofish.tls_handshake_timeout", time.Second*10)
	viper.SetDefault("bmc.gofish.idle_conn_timeout", time.Second*90)
}
