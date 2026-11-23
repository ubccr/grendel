// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"strings"

	"github.com/stmcginnis/gofish/oem/dell"
	"github.com/stmcginnis/gofish/schemas"
)

func (c *conn) isDell() bool {
	return strings.EqualFold(c.service.Vendor, "Dell")
}

// serialNumber returns what an operator would call this machine's serial. On an iDRAC that is the service tag, which is in SKU
func (c *conn) serialNumber(sys *schemas.ComputerSystem) string {
	if c.isDell() && sys.SKU != "" {
		return sys.SKU
	}
	return sys.SerialNumber
}

// isSELByID matches the ids a manager's event log goes by on firmware that types no log as SEL: Sel on an iDRAC and Log1 on a Supermicro
func isSELByID(id string) bool {
	return strings.EqualFold(id, "Sel") || strings.EqualFold(id, "Log1")
}

// isDellFaultList excludes a log an iDRAC types as SEL even though it is not the event log, and which rejects ClearLog
func isDellFaultList(l *schemas.LogService) bool {
	return strings.EqualFold(l.ID, "FaultList")
}

func dellOEMSystem(sys *schemas.ComputerSystem) (dell.OEMSystem, error) {
	cs, err := dell.FromComputerSystem(sys)
	if err != nil {
		return dell.OEMSystem{}, err
	}
	return cs.OEMSystem, nil
}
