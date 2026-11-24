// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"time"
)

var (
	DHCPListen = register[string](Option{
		Key:      "dhcp.listen",
		Section:  "dhcp",
		Summary:  "Listen address",
		Help:     "Address and UDP port the DHCP server binds to.",
		Kind:     KindString,
		Validate: validAddr,
		Default:  "0.0.0.0:67",
		Required: true,
		Apply:    ApplyRestart,
		Flag:     "dhcp-listen",
	})
	DHCPLeaseTime = register[time.Duration](Option{
		Key:     "dhcp.lease_time",
		Section: "dhcp",
		Summary: "Lease time",
		Help:    "How long a host may keep its address before renewing.",
		Kind:    KindDuration,
		Default: 24 * time.Hour,
		Apply:   ApplyRestart,
		Flag:    "dhcp-lease-time",
	})
	DHCPProxyOnly = register[bool](Option{
		Key:     "dhcp.proxy_only",
		Section: "dhcp",
		Summary: "Proxy only",
		Help:    "Only answer PXE requests and leave handing out addresses to another DHCP server.",
		Kind:    KindBool,
		Default: false,
		Apply:   ApplyRestart,
		Flag:    "dhcp-proxy-only",
	})
	DHCPDNSServers = register[[]string](Option{
		Key:      "dhcp.dns_servers",
		Section:  "dhcp",
		Summary:  "DNS servers",
		Help:     "DNS servers handed to hosts that are in none of the subnets below.",
		Kind:     KindStringList,
		Default:  []string{},
		Apply:    ApplyLive,
		Flag:     "dhcp-dns-servers",
		Validate: validateIPs,
	})
	DHCPDomainSearch = register[[]string](Option{
		Key:     "dhcp.domain_search",
		Section: "dhcp",
		Summary: "Search domains",
		Help:    "DNS search domains handed to hosts that are in none of the subnets below.",
		Kind:    KindStringList,
		Default: []string{},
		Apply:   ApplyLive,
		Flag:    "dhcp-domain-search",
	})
	DHCPMTU = register[int](Option{
		Key:     "dhcp.mtu",
		Section: "dhcp",
		Summary: "MTU",
		Help:    "Interface MTU handed to hosts that are in none of the subnets below.",
		Kind:    KindInt,
		Default: 1500,
		Min:     ptr(68),
		Max:     ptr(65535),
		Apply:   ApplyLive,
		Flag:    "dhcp-mtu",
	})
	DHCPGateway = register[string](Option{
		Key:      "dhcp.gateway",
		Section:  "dhcp",
		Group:    "Gateway",
		Summary:  "Default gateway",
		Help:     "One gateway handed to every host that is in none of the subnets below. Leave empty to hand out none.",
		Kind:     KindString,
		Validate: validIP,
		Default:  "",
		Apply:    ApplyLive,
		Flag:     "dhcp-gateway",
	})
	DHCPRouterOctet4 = register[int](Option{
		Key:     "dhcp.router_octet4",
		Section: "dhcp",
		Group:   "Gateway",
		Summary: "Gateway last octet",
		Help:    "Work out each host's gateway from its own address by replacing the last octet. With 254, a host at 10.104.13.10 is given 10.104.13.254. 0 turns this off.",
		Kind:    KindInt,
		Default: 0,
		Min:     ptr(0),
		Max:     ptr(255),
		Apply:   ApplyLive,
		Flag:    "dhcp-router-octet4",
	})
	DHCPRouter = register[string](Option{
		Key:      "dhcp.router",
		Section:  "dhcp",
		Group:    "Gateway",
		Summary:  "Fallback router",
		Help:     "A router address used when neither a subnet, the default gateway nor the gateway last octet gives one.",
		Kind:     KindString,
		Validate: validIP,
		Default:  "",
		Apply:    ApplyLive,
	})
	DHCPSubnets = register[map[string]any](Option{
		Key:        "dhcp.subnets",
		Section:    "dhcp",
		Group:      "Subnets",
		Summary:    "Subnets",
		Help:       "Per-subnet options, each under a name of your choosing. A host is handed the gateway, DNS servers, search domains and MTU of the first subnet its address falls inside, in place of the defaults above. Subnets are tried in name order, so 10-storage comes before 20-compute, and a subnet written without a name is numbered 10, 20, 30 by its position.",
		Kind:       KindGroup,
		EntryLabel: "Subnet",
		Default:    map[string]any{},
		Apply:      ApplyLive,
		Fields: []Field{
			{Name: "gateway", Summary: "Gateway", Kind: KindString, Required: true, Validate: validPrefix, Help: "The gateway's address with the subnet's prefix length, such as 10.17.41.254/23. It sets both the gateway and which hosts the entry applies to."},
			{Name: "dns", Summary: "DNS servers", Kind: KindStringList, Validate: validateIPs, Help: "DNS servers for the subnet."},
			{Name: "domainsearch", Summary: "Search domains", Kind: KindStringList, Help: "Search domains for the subnet."},
			{Name: "mtu", Summary: "MTU", Kind: KindInt, Validate: validMTU, Help: "Interface MTU for the subnet."},
		},
	})
	DHCPNetmask = register[int](Option{
		Key:        "dhcp.netmask",
		Section:    "dhcp",
		Summary:    "Netmask",
		Help:       "Has no effect.",
		Kind:       KindInt,
		Default:    0,
		Apply:      ApplyLive,
		Flag:       "dhcp-netmask",
		Deprecated: "it has no effect",
	})
)

func validateIPs(v any) error {
	for _, s := range v.([]string) {
		if _, err := netip.ParseAddr(s); err != nil {
			return fmt.Errorf("%s is not an IP address", s)
		}
	}
	return nil
}

func validPrefix(v any) error {
	if _, err := netip.ParsePrefix(v.(string)); err != nil {
		return errors.New("must be an address with a prefix length, such as 10.17.41.254/23")
	}
	return nil
}

func validMTU(v any) error {
	if mtu := v.(int); mtu < 0 || mtu > 65535 {
		return errors.New("must be between 0 and 65535")
	}
	return nil
}

// validAddr and validIP accept an empty value, which an option that needs one refuses by being Required
func validAddr(v any) error {
	if s := v.(string); s != "" {
		if _, _, err := net.SplitHostPort(s); err != nil {
			return errors.New("must be an address and port such as 0.0.0.0:53")
		}
	}
	return nil
}

func validIP(v any) error {
	if s := v.(string); s != "" {
		if _, err := netip.ParseAddr(s); err != nil {
			return errors.New("must be an IP address")
		}
	}
	return nil
}
