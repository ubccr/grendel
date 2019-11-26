package util

import (
	"errors"
	"fmt"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/interfaces"
)

func GetInterfaceIP() (net.IP, error) {
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
