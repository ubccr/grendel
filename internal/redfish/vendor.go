// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"slices"
	"strings"

	"github.com/stmcginnis/gofish/oem/dell"
	"github.com/stmcginnis/gofish/schemas"
)

// isDell reports whether this is an iDRAC
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

// knownSystemLogIDs name the log an operator means by the system event log, on vendors that do not type it as SEL. HPE calls it IML, Supermicro Log1, Lenovo StandardLog and OpenBMC EventLog. These apply only under Systems, because the same ids can mean something else under Managers
var knownSystemLogIDs = []string{"iml", "log1", "standardlog", "eventlog"}

func isKnownSystemLogID(id string) bool {
	return slices.Contains(knownSystemLogIDs, strings.ToLower(id))
}

// isDellFaultList excludes a log an iDRAC types as SEL even though it is not the event log, and which rejects ClearLog
func isDellFaultList(l *schemas.LogService) bool {
	return strings.EqualFold(l.ID, "FaultList")
}

// dellOEMSystem pulls the Dell OEM block out of a computer system
func dellOEMSystem(sys *schemas.ComputerSystem) (dell.OEMSystem, error) {
	cs, err := dell.FromComputerSystem(sys)
	if err != nil {
		return dell.OEMSystem{}, err
	}
	return cs.OEMSystem, nil
}
