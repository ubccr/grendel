// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package sqlstore

import "github.com/spf13/viper"

const defaultMaxReadConns = 8
const defaultDriver = "sqlite3"

func init() {
	viper.SetDefault("database.max_read_conns", defaultMaxReadConns)
}
