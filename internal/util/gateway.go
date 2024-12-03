// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package util

import (
	"net"

	"github.com/spf13/viper"
)

func DefaultGateway(ip net.IP) net.IP {
	var router net.IP
	if viper.IsSet("dhcp.router_octet4") {
		router = ip.Mask(net.CIDRMask(24, 32))
		router[3] += byte(viper.GetInt("dhcp.router_octet4"))
	} else if viper.IsSet("dhcp.router") {
		router = net.ParseIP(viper.GetString("dhcp.router"))
	}

	return router
}
