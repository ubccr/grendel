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

package util

import (
	"errors"
	"fmt"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/interfaces"
)

func GetFirstExternalIPFromInterfaces() (net.IP, error) {
	// Attempt to discover IP from interfaces
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

	if len(serverIps) != 1 {
		//TODO add support for multiple interfaces
		return nil, fmt.Errorf("Multiple interfaces not supported yet: %#v", serverIps)
	}

	return serverIps[0], nil
}

func GetInterfaceFromIP(ip net.IP) (string, error) {
	// Attempt to discover interface from IP
	intfs, err := interfaces.GetNonLoopbackInterfaces()
	if err != nil {
		return "", err
	}

	for _, intf := range intfs {
		addrs, err := intf.Addrs()
		if err != nil {
			return "", err
		}

		ips, err := dhcpv4.GetExternalIPv4Addrs(addrs)
		if err != nil {
			return "", err
		}

		if len(ips) == 0 {
			continue
		}

		for _, i := range ips {
			if i.To4() != nil && i.To4().Equal(ip) {
				return intf.Name, nil
			}
		}
	}

	return "", fmt.Errorf("Interface not found with ip: %s", ip)
}
