// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/internal/dns"
)

func init() {
	serveCmd.PersistentFlags().String("dns-listen", "0.0.0.0:53", "address to listen on")
	serveCmd.PersistentFlags().Int("dns-ttl", 300, "ttl for dns records")
	serveCmd.PersistentFlags().String("dns-forward", "", "address to forward dns queries not resolved by grendel. ex: 1.1.1.1:53")
	viper.BindPFlag("dns.listen", serveCmd.PersistentFlags().Lookup("dns-listen"))
	viper.BindPFlag("dns.ttl", serveCmd.PersistentFlags().Lookup("dns-ttl"))
	viper.BindPFlag("dns.forward", serveCmd.PersistentFlags().Lookup("dns-forward"))

	register(&Service{
		Name: "dns",
		New:  newDNS,
	})
}

func newDNS() (Runner, error) {
	srv, err := dns.NewServer(DB, getListenAddress("dns.listen"), viper.GetInt("dns.ttl"))
	if err != nil {
		return nil, err
	}

	if fwAddr := viper.GetString("dns.forward"); fwAddr != "" {
		shared.Log.Debugf("dns.forward address set, using: %s", fwAddr)
	}

	return srv, nil
}
