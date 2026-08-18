// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"time"

	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/dhcp"
	"github.com/ubccr/grendel/internal/logger"
)

var dhcpLog = logger.GetLogger("DHCP")

func init() {
	serveCmd.PersistentFlags().String("dhcp-listen", "0.0.0.0:67", "address to listen on")
	viper.BindPFlag("dhcp.listen", serveCmd.PersistentFlags().Lookup("dhcp-listen"))
	serveCmd.PersistentFlags().String("dhcp-lease-time", "24h", "default lease time")
	viper.BindPFlag("dhcp.lease_time", serveCmd.PersistentFlags().Lookup("dhcp-lease-time"))
	serveCmd.PersistentFlags().StringSlice("dhcp-dns-servers", []string{}, "dns servers list")
	viper.BindPFlag("dhcp.dns_servers", serveCmd.PersistentFlags().Lookup("dhcp-dns-servers"))
	serveCmd.PersistentFlags().StringSlice("dhcp-domain-search", []string{}, "domain name search list")
	viper.BindPFlag("dhcp.domain_search", serveCmd.PersistentFlags().Lookup("dhcp-domain-search"))
	serveCmd.PersistentFlags().Int("dhcp-mtu", 1500, "default mtu")
	viper.BindPFlag("dhcp.mtu", serveCmd.PersistentFlags().Lookup("dhcp-mtu"))
	serveCmd.PersistentFlags().Bool("dhcp-proxy-only", false, "only run boot proxy")
	viper.BindPFlag("dhcp.proxy_only", serveCmd.PersistentFlags().Lookup("dhcp-proxy-only"))
	serveCmd.PersistentFlags().Int("dhcp-router-octet4", 0, "automatic router configuration")
	viper.BindPFlag("dhcp.router_octet4", serveCmd.PersistentFlags().Lookup("dhcp-router-octet4"))
	serveCmd.PersistentFlags().String("dhcp-gateway", "", "static gateway address")
	viper.BindPFlag("dhcp.gateway", serveCmd.PersistentFlags().Lookup("dhcp-gateway"))
	serveCmd.PersistentFlags().Int("dhcp-netmask", 0, "subnet mask")
	viper.BindPFlag("dhcp.netmask", serveCmd.PersistentFlags().Lookup("dhcp-netmask"))

	register(&Service{
		Name: "dhcp",
		New:  newDHCP,
	})
}

func newDHCP() (Runner, error) {
	srv, err := dhcp.NewServer(DB, getListenAddress("dhcp.listen"))
	if err != nil {
		return nil, err
	}

	leaseTime, err := time.ParseDuration(viper.GetString("dhcp.lease_time"))
	if err != nil {
		return nil, err
	}

	srv.LeaseTime = leaseTime
	dhcpLog.Infof("Default lease time: %s", srv.LeaseTime)

	srv.ProxyOnly = viper.GetBool("dhcp.proxy_only")
	if srv.ProxyOnly {
		dhcpLog.Infof("Running in ProxyOnly mode")
	}

	return srv, nil
}
