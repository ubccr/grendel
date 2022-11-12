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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/dhcp"
	"github.com/ubccr/grendel/logger"
	"gopkg.in/tomb.v2"
)

func init() {
	dhcpCmd.PersistentFlags().String("dhcp-listen", "0.0.0.0:67", "address to listen on")
	viper.BindPFlag("dhcp.listen", dhcpCmd.PersistentFlags().Lookup("dhcp-listen"))
	dhcpCmd.PersistentFlags().String("dhcp-lease-time", "24h", "default lease time")
	viper.BindPFlag("dhcp.lease_time", dhcpCmd.PersistentFlags().Lookup("dhcp-lease-time"))
	dhcpCmd.PersistentFlags().StringSlice("dhcp-dns-servers", []string{}, "dns servers list")
	viper.BindPFlag("dhcp.dns_servers", dhcpCmd.PersistentFlags().Lookup("dhcp-dns-servers"))
	dhcpCmd.PersistentFlags().StringSlice("dhcp-domain-search", []string{}, "domain name search list")
	viper.BindPFlag("dhcp.domain_search", dhcpCmd.PersistentFlags().Lookup("dhcp-domain-search"))
	dhcpCmd.PersistentFlags().Int("dhcp-mtu", 1500, "default mtu")
	viper.BindPFlag("dhcp.mtu", dhcpCmd.PersistentFlags().Lookup("dhcp-mtu"))
	dhcpCmd.PersistentFlags().Bool("dhcp-proxy-only", false, "only run boot proxy")
	viper.BindPFlag("dhcp.proxy_only", dhcpCmd.PersistentFlags().Lookup("dhcp-proxy-only"))
	dhcpCmd.PersistentFlags().Int("dhcp-router-octet4", 0, "automatic router configuration")
	viper.BindPFlag("dhcp.router_octet4", dhcpCmd.PersistentFlags().Lookup("dhcp-router-octet4"))
	dhcpCmd.PersistentFlags().String("dhcp-gateway", "", "static gateway address")
	viper.BindPFlag("dhcp.gateway", dhcpCmd.PersistentFlags().Lookup("dhcp-gateway"))
	dhcpCmd.PersistentFlags().Int("dhcp-netmask", 0, "subnet mask")
	viper.BindPFlag("dhcp.netmask", dhcpCmd.PersistentFlags().Lookup("dhcp-netmask"))

	serveCmd.AddCommand(dhcpCmd)
}

var (
	dhcpLog = logger.GetLogger("DHCP")
	dhcpCmd = &cobra.Command{
		Use:   "dhcp",
		Short: "Run DHCP server",
		Long:  `Run DHCP server`,
		RunE: func(command *cobra.Command, args []string) error {
			t := NewInterruptTomb()
			t.Go(func() error { return serveDHCP(t) })
			return t.Wait()
		},
	}
)

func serveDHCP(t *tomb.Tomb) error {
	dhcpListen, err := GetListenAddress(viper.GetString("dhcp.listen"))
	if err != nil {
		return err
	}

	srv, err := dhcp.NewServer(DB, dhcpListen)
	if err != nil {
		return err
	}

	leaseTime, err := time.ParseDuration(viper.GetString("dhcp.lease_time"))
	if err != nil {
		return err
	}

	srv.LeaseTime = leaseTime
	dhcpLog.Infof("Default lease time: %s", srv.LeaseTime)

	srv.ProxyOnly = viper.GetBool("dhcp.proxy_only")
	if srv.ProxyOnly {
		dhcpLog.Infof("Running in ProxyOnly mode")
	}

	t.Go(srv.Serve)
	t.Go(func() error {
		time.Sleep(1 * time.Second)
		<-t.Dying()
		dhcpLog.Info("Shutting down DHCP server...")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctxShutdown); err != nil {
			dhcpLog.Errorf("Failed shutting down DHCP server: %s", err)
			return err
		}

		return nil
	})

	return nil
}
