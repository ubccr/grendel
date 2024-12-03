// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package util

import (
	"errors"
	"fmt"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/interfaces"
)

func GetFirstExternalIPFromInterfaces() (net.IP, error) {
	intfs, err := interfaces.GetNonLoopbackInterfaces()
	if err != nil {
		return nil, err
	}

	serverIps := make([]net.IP, 0)

	for _, intf := range intfs {
		addrs, err := intf.Addrs()
		if err != nil {
			return nil, err
		}

		ips, err := dhcpv4.GetExternalIPv4Addrs(addrs)
		if err != nil {
			return nil, err
		}

		if len(ips) == 0 {
			continue
		}

		serverIps = append(serverIps, ips...)
	}

	if len(serverIps) == 0 {
		return nil, errors.New("Failed to find server ip address from configured interfaces")
	}

	// Multiple interfaces found. Using first one?
	// This is used for setting the ServerIP in dhcp responses etc.
	return serverIps[0], nil
}

func GetInterfaceIPMap() (map[int]net.IP, error) {
	intfs, err := interfaces.GetNonLoopbackInterfaces()
	if err != nil {
		return nil, err
	}

	intfIps := make(map[int]net.IP, 0)

	for _, intf := range intfs {
		addrs, err := intf.Addrs()
		if err != nil {
			return nil, err
		}

		ips, err := dhcpv4.GetExternalIPv4Addrs(addrs)
		if err != nil {
			return nil, err
		}

		if len(ips) == 0 {
			continue
		}

		// XXX support interfaces with multiple IPs?
		// This is used for setting the ServerIP in dhcp responses so we just
		// use the first one for now.
		intfIps[intf.Index] = ips[0]
	}

	return intfIps, nil
}

func GetInterfaceFromIP(ip net.IP) (string, net.IPMask, error) {
	intfs, err := interfaces.GetNonLoopbackInterfaces()
	if err != nil {
		return "", nil, err
	}

	for _, intf := range intfs {
		addrs, err := intf.Addrs()
		if err != nil {
			return "", nil, err
		}

		for _, addr := range addrs {
			var i net.IP
			var mask net.IPMask
			switch v := addr.(type) {
			case *net.IPAddr:
				i = v.IP
			case *net.IPNet:
				i = v.IP
				mask = v.Mask
			}

			if i.To4() != nil && i.To4().Equal(ip) {
				return intf.Name, mask, nil
			}
		}

	}

	return "", nil, fmt.Errorf("Interface not found with ip: %s", ip)
}
