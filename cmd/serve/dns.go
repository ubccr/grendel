// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package serve

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/dns"
)

func init() {
	dnsCmd.PersistentFlags().String("dns-listen", "0.0.0.0:53", "address to listen on")
	dnsCmd.PersistentFlags().Int("dns-ttl", 300, "ttl for dns records")
	viper.BindPFlag("dns.listen", dnsCmd.PersistentFlags().Lookup("dns-listen"))
	viper.BindPFlag("dns.ttl", dnsCmd.PersistentFlags().Lookup("dns-ttl"))

	serveCmd.AddCommand(dnsCmd)
}

var (
	dnsCmd = &cobra.Command{
		Use:   "dns",
		Short: "Run DNS server",
		Long:  `Run DNS server`,
		RunE: func(command *cobra.Command, args []string) error {
			ctx, _ := NewInterruptContext()
			return serveDNS(ctx)
		},
	}
)

func serveDNS(ctx context.Context) error {
	dnsServer, err := dns.NewServer(DB, viper.GetString("dns.listen"), viper.GetInt("dns.ttl"))
	if err != nil {
		return err
	}

	if err := dnsServer.Serve(ctx); err != nil {
		return err
	}

	return nil
}
