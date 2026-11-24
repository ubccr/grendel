// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"github.com/ubccr/grendel/internal/dhcp"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/settings"
)

var dhcpLog = logger.GetLogger("DHCP")

func init() {

	register(&Service{
		Name: "dhcp",
		New:  newDHCP,
	})
}

func newDHCP() (Runner, error) {
	srv, err := dhcp.NewServer(DB, getListenAddress("dhcp.listen"))
	if err != nil {
		return nil, err
	}

	srv.LeaseTime = settings.DHCPLeaseTime.Get()
	dhcpLog.Infof("Default lease time: %s", srv.LeaseTime)

	srv.ProxyOnly = settings.DHCPProxyOnly.Get()
	if srv.ProxyOnly {
		dhcpLog.Infof("Running in ProxyOnly mode")
	}

	return srv, nil
}
