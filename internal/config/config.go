// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"maps"
	"net"
	"net/netip"
	"slices"

	"github.com/ubccr/grendel/internal/settings"
)

type Subnet struct {
	Gateway      netip.Prefix
	DNS          []net.IP
	DomainSearch []string
	MTU          uint16
}

type network struct {
	provisionAddr netip.AddrPort
	subnets       []Subnet
	defaultDNS    []net.IP
	gateway       netip.Addr
}

// parsed is built again whenever a setting changes. The settings were validated when they were set, so anything that still fails to parse is left out.
var parsed = settings.Derive(func() network {
	n := network{subnets: []Subnet{}, defaultDNS: []net.IP{}}

	// the first subnet holding a host's address is the one used, and they are tried in name order: a natural one, so 2-lab comes before 10-core
	subnets := settings.DHCPSubnets.Get()
	names := slices.Collect(maps.Keys(subnets))
	slices.SortFunc(names, settings.NaturalCompare)
	for _, name := range names {
		row, _ := subnets[name].(map[string]any)
		gateway, _ := row["gateway"].(string)
		gw, err := netip.ParsePrefix(gateway)
		if err != nil {
			continue
		}
		dns, _ := row["dns"].([]string)
		domains, _ := row["domainsearch"].([]string)
		mtu, _ := row["mtu"].(int)
		n.subnets = append(n.subnets, Subnet{Gateway: gw, DNS: ips(dns), DomainSearch: domains, MTU: uint16(mtu)})
	}

	n.defaultDNS = ips(settings.DHCPDNSServers.Get())
	n.provisionAddr, _ = netip.ParseAddrPort(settings.ProvisionListen.Get())
	if gw := settings.DHCPGateway.Get(); gw != "" {
		n.gateway, _ = netip.ParseAddr(gw)
	}
	return n
})

func ips(texts []string) []net.IP {
	out := make([]net.IP, 0, len(texts))
	for _, text := range texts {
		if addr, err := netip.ParseAddr(text); err == nil {
			out = append(out, net.IP(addr.AsSlice()))
		}
	}
	return out
}

func ProvisionAddr() netip.AddrPort { return parsed.Get().provisionAddr }

func ProvisionHostname() string { return settings.ProvisionHostname.Get() }

func ProvisionScheme() string {
	if settings.ProvisionCert.Get() != "" && settings.ProvisionKey.Get() != "" {
		return "https"
	}
	return "http"
}

func Subnets() []Subnet { return parsed.Get().subnets }

func DefaultDNS() []net.IP { return parsed.Get().defaultDNS }

func DefaultDomainSearch() []string { return settings.DHCPDomainSearch.Get() }

func DefaultMTU() uint16 { return uint16(settings.DHCPMTU.Get()) }

// DefaultGateway is the zero Addr when no gateway is set
func DefaultGateway() netip.Addr { return parsed.Get().gateway }
