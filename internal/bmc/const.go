// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import "github.com/spf13/viper"

const (
	delay  = 1
	fanout = 5
)

func init() {
	viper.SetDefault("bmc.delay", delay)
	viper.SetDefault("bmc.fanout", fanout)
}
