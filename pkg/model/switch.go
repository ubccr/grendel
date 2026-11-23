// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"net"
)

type MACTableEntry struct {
	Ifname string           `json:"ifname"`
	VLAN   string           `json:"vlan"`
	Type   string           `json:"type"`
	MAC    net.HardwareAddr `json:"mac"`
}

type LLDP struct {
	PortName          string `json:"port_name"`
	ChassisIdType     string `json:"chassis_id_type"`
	ChassisId         string `json:"chassis_id"`
	SystemName        string `json:"system_name"`
	SystemDescription string `json:"system_description"`
	ManagementAddress string `json:"management_address"`
	PortDescription   string `json:"port_description"`
	PortId            string `json:"port_id"`
	PortIdType        string `json:"port_id_type"`
}

type MACTable []*MACTableEntry

type LLDPNeighbors map[string]*LLDP

type SwitchLLDPList []LLDP
