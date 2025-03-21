// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/config"
	"go4.org/netipx"
)

// Enum for NicType
type NicType int

var ErrInvalidNicType = errors.New("Invalid nic type")

const (
	NicTypeEthernet NicType = iota + 1
	NicTypeBMC
	NicTypeBond
)

type NetInterfaceList []NetInterface

type NetInterface struct {
	ID   int64            `json:"id"`
	MAC  net.HardwareAddr `json:"mac"`
	Name string           `json:"ifname"`
	IP   netip.Prefix     `json:"ip"`
	FQDN string           `json:"fqdn"`
	BMC  bool             `json:"bmc"`
	VLAN string           `json:"vlan"`
	MTU  uint16           `json:"mtu,omitempty"`
}

// Return the string of a NicType
func (n NicType) String() string {
	switch n {
	case NicTypeEthernet:
		return "ethernet"
	case NicTypeBMC:
		return "bmc"
	case NicTypeBond:
		return "bond"
	default:
		return "Unknown type"
	}
}

// New NicType from string
func NicTypeFromString(name string) (NicType, error) {
	switch name {
	case "ethernet":
		return NicTypeEthernet, nil
	case "bmc":
		return NicTypeBMC, nil
	case "bond":
		return NicTypeBond, nil
	default:
		return NicTypeEthernet, ErrInvalidNicType
	}
}

func (n *NetInterfaceList) Scan(value interface{}) error {
	data, ok := value.(string)
	if !ok {
		return errors.New("incompatible type")
	}
	var nlist NetInterfaceList
	err := json.Unmarshal([]byte(data), &nlist)
	if err != nil {
		return fmt.Errorf("failed to decode: %w", err)
	}

	*n = nlist
	return nil
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

	for _, subnet := range config.Subnets {
		if subnet.MTU == 0 {
			continue
		}

		if subnet.Gateway.Contains(n.IP.Addr()) {
			return subnet.MTU
		}
	}

	return config.DefaultMTU
}

func (n *NetInterface) Gateway() netip.Addr {
	for _, subnet := range config.Subnets {
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

	return config.DefaultGateway
}

func (n *NetInterface) DNS() []net.IP {
	dnsServers := make([]net.IP, 0)

	for _, subnet := range config.Subnets {
		if len(subnet.DNS) == 0 {
			continue
		}

		if subnet.Gateway.Contains(n.IP.Addr()) {
			dnsServers = append(dnsServers, subnet.DNS...)
			break
		}
	}

	if len(dnsServers) == 0 {
		dnsServers = append(dnsServers, config.DefaultDNS...)
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
	for _, subnet := range config.Subnets {
		if len(subnet.DomainSearch) == 0 {
			continue
		}

		if subnet.Gateway.Contains(n.IP.Addr()) {
			return subnet.DomainSearch
		}
	}

	return config.DefaultDomainSearch
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
