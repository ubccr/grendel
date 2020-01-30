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

package tors

import (
	"encoding/json"
	"net"

	"github.com/ubccr/grendel/logger"
)

var log = logger.GetLogger("SWITCH")

type MACTableEntry struct {
	Ifname string           `json:"ifname"`
	Port   int              `json:"port"`
	VLAN   string           `json:"vlan"`
	Type   string           `json:"type"`
	MAC    net.HardwareAddr `json:"mac-addr"`
}

type MACTable map[string]*MACTableEntry

type NetworkSwitch interface {
	GetMACTable() (MACTable, error)
}

func (mt MACTable) Port(port int) []*MACTableEntry {
	entries := make([]*MACTableEntry, 0)
	for _, entry := range mt {
		if entry.Port == port {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (mt MACTable) String() string {
	data, _ := json.MarshalIndent(mt, "", "    ")
	return string(data)
}
