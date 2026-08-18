// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package dns

import (
	"time"

	"github.com/spf13/viper"
)

// defaultQueryTimeout matches the miekg/dns client default. Past this the
// client has already given up, so any work still holding a database connection
// is wasted.
const defaultQueryTimeout = 2 * time.Second

func init() {
	viper.SetDefault("dns.query_timeout", defaultQueryTimeout)
}
