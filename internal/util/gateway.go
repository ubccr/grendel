// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package util

import (
	"net"

	"github.com/ubccr/grendel/internal/settings"
)

func DefaultGateway(ip net.IP) net.IP {
	var router net.IP
	if octet := settings.DHCPRouterOctet4.Get(); octet > 0 {
		router = ip.Mask(net.CIDRMask(24, 32))
		router[3] += byte(octet)
	} else if addr := settings.DHCPRouter.Get(); addr != "" {
		router = net.ParseIP(addr)
	}

	return router
}
