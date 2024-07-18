// Copyright 2019 Grendel Authors. All rights reserved.  //
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

package model

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/spf13/viper"
	"go4.org/netipx"
)

type Subnet struct {
	Gateway      netip.Prefix
	DNS          []net.IP
	DomainSearch []string
	MTU          uint16
}

type NetInterface struct {
	HostID ksuid.KSUID
	ID     uint16           `gorm:"primaryKey"`
	MAC    net.HardwareAddr `json:"mac" validate:"required" gorm:"serializer:MACSerializer"`
	Name   string           `json:"ifname"`
	IP     netip.Prefix     `json:"ip" gorm:"serializer:IPPrefixSerializer"`
	FQDN   string           `json:"fqdn"`
	BMC    bool             `json:"bmc"`
	VLAN   string           `json:"vlan"`
	MTU    uint16           `json:"mtu,omitempty"`
}

func (n *NetInterface) MarshalJSON() ([]byte, error) {
	type Alias NetInterface
	return json.Marshal(&struct {
		MAC string `json:"mac"`
		IP  string `json:"ip"`
		*Alias
	}{
		MAC:   n.MAC.String(),
		IP:    n.CIDR(),
		Alias: (*Alias)(n),
	})
}

func (n *NetInterface) UnmarshalJSON(data []byte) error {
	type Alias NetInterface
	aux := &struct {
		MAC string `json:"mac"`
		IP  string `json:"ip"`
		*Alias
	}{
		Alias: (*Alias)(n),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.MAC != "" {
		mac, err := net.ParseMAC(aux.MAC)
		if err != nil {
			return err
		}

		n.MAC = mac
	}

	if aux.IP != "" {
		ip, err := netip.ParsePrefix(aux.IP)
		if err != nil {
			return fmt.Errorf("Invalid IPv4 address %s: %s", aux.IP, err)
		}

		n.IP = ip
	}
	return nil
}

func (n *NetInterface) CIDR() string {
	if !n.IP.IsValid() {
		return ""
	}

	return n.IP.String()
}

func (n *NetInterface) AddrString() string {
	if !n.IP.IsValid() {
		return ""
	}

	if !n.IP.Addr().IsValid() {
		return ""
	}

	return n.IP.Addr().String()
}

func (n *NetInterface) Addr() netip.Addr {
	return n.IP.Addr()
}

func (n *NetInterface) ToStdAddr() net.IP {
	if !n.IP.IsValid() {
		return net.ParseIP("")
	}

	addr := n.IP.Addr()

	if addr.Is4In6() {
		addr = addr.Unmap()
	}

	return net.IP(addr.AsSlice())
}

func (n *NetInterface) Netmask() net.IPMask {
	return net.CIDRMask(n.IP.Bits(), 32)
}

func (n *NetInterface) NetmaskString() string {
	return net.IP(net.CIDRMask(n.IP.Bits(), 32)).String()
}

func (n *NetInterface) InterfaceMTU() uint16 {
	if n.MTU != 0 {
		return n.MTU
	}

	for _, subnet := range Subnets {
		if subnet.MTU == 0 {
			continue
		}

		if subnet.Gateway.Contains(n.IP.Addr()) {
			return subnet.MTU
		}
	}

	return DefaultMTU
}

func (n *NetInterface) Gateway() netip.Addr {
	for _, subnet := range Subnets {
		if subnet.Gateway.Contains(n.IP.Addr()) {
			return subnet.Gateway.Addr()
		}
	}

	if viper.IsSet("dhcp.router_octet4") {
		lastIP := netipx.PrefixLastIP(n.IP)
		ip4 := lastIP.As4()
		ip4[3] = uint8(viper.GetInt("dhcp.router_octet4"))
		return netip.AddrFrom4(ip4)
	}

	return DefaultGateway
}

func (n *NetInterface) DNS() []net.IP {
	dnsServers := make([]net.IP, 0)

	for _, subnet := range Subnets {
		if len(subnet.DNS) == 0 {
			continue
		}

		if subnet.Gateway.Contains(n.IP.Addr()) {
			dnsServers = append(dnsServers, subnet.DNS...)
			break
		}
	}

	if len(dnsServers) == 0 {
		dnsServers = append(dnsServers, DefaultDNS...)
	}

	if len(dnsServers) > 1 {
		// Randomize DNS servers to distribute load
		// TODO: add option to turn this off
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(dnsServers), func(i, j int) { dnsServers[i], dnsServers[j] = dnsServers[j], dnsServers[i] })
	}

	return dnsServers
}

func (n *NetInterface) DomainSearch() []string {
	for _, subnet := range Subnets {
		if len(subnet.DomainSearch) == 0 {
			continue
		}

		if subnet.Gateway.Contains(n.IP.Addr()) {
			return subnet.DomainSearch
		}
	}

	return DefaultDomainSearch
}

func (n *NetInterface) DNSList() []string {
	dnsServers := n.DNS()
	dnsList := make([]string, len(dnsServers))

	for i, dip := range dnsServers {
		dnsList[i] = dip.String()
	}

	return dnsList
}

func (n *NetInterface) HostNameIndex(idx int) string {
	names := strings.Split(n.FQDN, ",")
	if idx >= 0 && idx < len(names) {
		return names[idx]
	}

	return ""
}

func (n *NetInterface) HostName() string {
	names := strings.Split(n.FQDN, ",")
	return names[0]
}

func (n *NetInterface) ShortName() string {
	parts := strings.Split(n.HostName(), ".")
	return parts[0]
}

func (n *NetInterface) Domain() string {
	parts := strings.Split(n.HostName(), ".")
	if len(parts) > 1 {
		return strings.Join(parts[1:], ".")
	}

	return parts[0]
}
