// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/spf13/viper"
)

var (
	ProvisionAddr       netip.AddrPort = netip.MustParseAddrPort("0.0.0.0:80")
	ProvisionScheme     string         = "http"
	ProvisionHostname   string         = ""
	Subnets             []Subnet       = []Subnet{}
	DefaultDNS          []net.IP       = []net.IP{}
	DefaultDomainSearch []string       = []string{}
	DefaultMTU          uint16         = 1500
	DefaultGateway      netip.Addr
)

func ParseConfigs() error {
	type SubnetConfig struct {
		Gateway      string
		DNS          string
		DomainSearch string
		MTU          uint16
	}
	var subnetConfigs []SubnetConfig

	err := viper.UnmarshalKey("dhcp.subnets", &subnetConfigs)
	if err != nil {
		return err
	}

	Subnets = make([]Subnet, 0)
	for _, sc := range subnetConfigs {
		gw, err := netip.ParsePrefix(sc.Gateway)
		if err != nil {
			return fmt.Errorf("Failed parsing dhcp.subnets config. Invalid gateway: %s", sc.Gateway)
		}
		dnsServers := make([]net.IP, 0)
		for _, dnsIP := range strings.Split(sc.DNS, ",") {
			if dnsIP == "" {
				continue
			}

			d, err := netip.ParseAddr(dnsIP)
			if err != nil {
				return fmt.Errorf("Failed parsing dhcp.subnets config. Invalid dns: %s", dnsIP)
			}
			dnsServers = append(dnsServers, net.IP(d.AsSlice()))
		}

		domainSearch := make([]string, 0)
		for _, domain := range strings.Split(sc.DomainSearch, ",") {
			if domain == "" {
				continue
			}
			domainSearch = append(domainSearch, domain)
		}

		Subnets = append(Subnets, Subnet{Gateway: gw, DNS: dnsServers, DomainSearch: domainSearch, MTU: sc.MTU})
	}

	DefaultDNS = make([]net.IP, 0)
	for _, dnsIP := range viper.GetStringSlice("dhcp.dns_servers") {
		d, err := netip.ParseAddr(dnsIP)
		if err != nil {
			return fmt.Errorf("Failed parsing dhcp.dns_servers config. Invalid dns: %s", dnsIP)
		}
		DefaultDNS = append(DefaultDNS, net.IP(d.AsSlice()))
	}

	DefaultDomainSearch = viper.GetStringSlice("dhcp.domain_search")
	DefaultMTU = uint16(viper.GetInt("dhcp.mtu"))

	addrPort, err := netip.ParseAddrPort(viper.GetString("provision.listen"))
	if err != nil {
		return fmt.Errorf("Failed parsing provision.listen address %s: %w", viper.GetString("provision.listen"), err)
	}

	if viper.IsSet("dhcp.gateway") {
		DefaultGateway, err = netip.ParseAddr(viper.GetString("dhcp.gateway"))
		if err != nil {
			return fmt.Errorf("Failed parsing dhcp.gateway %s: %w", viper.GetString("dhcp.gateway"), err)
		}
	}

	ProvisionHostname = viper.GetString("provision.hostname")
	ProvisionAddr = addrPort

	if viper.IsSet("provision.cert") && viper.IsSet("provision.key") {
		ProvisionScheme = "https"
	}

	return nil
}
