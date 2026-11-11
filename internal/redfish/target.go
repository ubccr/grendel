// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"github.com/ubccr/grendel/pkg/model"
)

type Target struct {
	Name string
	Addr string
}

// TargetsFromHosts picks the hosts with an interface marked bmc
func TargetsFromHosts(hosts model.HostList) []Target {
	var targets []Target

	for _, host := range hosts {
		if host == nil {
			continue
		}

		nic := host.InterfaceBMC()
		if nic == nil {
			continue
		}

		targets = append(targets, Target{Name: host.Name, Addr: nic.AddrString()})
	}

	return targets
}
