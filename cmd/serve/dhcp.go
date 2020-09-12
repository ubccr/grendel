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
	"fmt"
	"net"
	"strconv"
	"strings"
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
	dhcpCmd.PersistentFlags().String("dhcp-router", "", "static router address")
	viper.BindPFlag("dhcp.router", dhcpCmd.PersistentFlags().Lookup("dhcp-router"))
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

	address, err := GetListenAddress(viper.GetString("provision.listen"))
	if err != nil {
		return err
	}

	ipStr, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	provisionIP := net.ParseIP(ipStr)
	if provisionIP == nil || provisionIP.To4() == nil {
		return fmt.Errorf("Invalid Provision IPv4 address: %s", ipStr)
	}

	if provisionIP.To4().Equal(net.IPv4zero) {
		// Assume we're running on same server as provision?
		provisionIP = srv.ServerAddress
	}

	srv.ProvisionPort, err = strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	srv.ProvisionScheme = viper.GetString("provision.scheme")
	if srv.ProvisionScheme == "" {
		srv.ProvisionScheme = "http"
	}

	if viper.IsSet("provision.hostname") {
		srv.ProvisionHostname = viper.GetString("provision.hostname")
	}

	if srv.ProvisionHostname == "" && srv.ProvisionScheme == "https" {
		hosts, err := net.LookupAddr(provisionIP.String())
		if err == nil && len(hosts) > 0 {
			fqdn := hosts[0]
			srv.ProvisionHostname = strings.TrimSuffix(fqdn, ".")
		}
	}

	if srv.ProvisionHostname == "" {
		srv.ProvisionHostname = provisionIP.String()
	}

	dhcpLog.Infof("Base URL for ipxe: %s://%s:%d", srv.ProvisionScheme, srv.ProvisionHostname, srv.ProvisionPort)

	if viper.IsSet("dhcp.dns_servers") {
		srv.DNSServers = make([]net.IP, 0)
		for _, arg := range viper.GetStringSlice("dhcp.dns_servers") {
			dnsip := net.ParseIP(arg)
			if dnsip == nil || dnsip.To4() == nil {
				return fmt.Errorf("Invalid dns server ip address: %s", arg)
			}
			srv.DNSServers = append(srv.DNSServers, dnsip)
		}
		dhcpLog.Infof("Using DNS servers: %v", srv.DNSServers)
	}

	if viper.IsSet("dhcp.domain_search") {
		srv.DomainSearchList = viper.GetStringSlice("dhcp.domain_search")
		dhcpLog.Infof("Using Domain Search List: %v", srv.DomainSearchList)
	}

	if viper.IsSet("dhcp.router") {
		routerIP := net.ParseIP(viper.GetString("dhcp.router"))
		if routerIP == nil || routerIP.To4() == nil {
			return fmt.Errorf("Invalid router ip address: %s", viper.GetString("dhcp.router"))
		}

		srv.RouterIP = routerIP
		dhcpLog.Infof("Static router: %s", srv.RouterIP)
	}

	leaseTime, err := time.ParseDuration(viper.GetString("dhcp.lease_time"))
	if err != nil {
		return err
	}

	srv.LeaseTime = leaseTime
	dhcpLog.Infof("Default lease time: %s", srv.LeaseTime)

	srv.MTU = viper.GetInt("dhcp.mtu")
	dhcpLog.Infof("Default mtu: %d", srv.MTU)

	srv.RouterOctet4 = viper.GetInt("dhcp.router_octet4")
	if srv.RouterOctet4 > 0 {
		srv.Netmask = net.CIDRMask(24, 32)
		dhcpLog.Infof("Using automatic router configuration")
		dhcpLog.Infof("Netmask: %s", srv.Netmask)
		dhcpLog.Infof("Router Octet4: %d", srv.RouterOctet4)
	} else if viper.GetInt("dhcp.netmask") > 0 {
		srv.Netmask = net.CIDRMask(viper.GetInt("dhcp.netmask"), 32)
		dhcpLog.Infof("Netmask: %s", srv.Netmask)
	}

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
