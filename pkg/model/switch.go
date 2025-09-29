package model

import (
	"encoding/json"
	"net"
)

type InterfaceStatus struct {
	Name                string           `json:"name"`
	LineProtocolStatus  string           `json:"line_protocol_status"`
	InterfaceStatus     string           `json:"interface_status"`
	PhysicalAddress     net.HardwareAddr `json:"physical_address"`
	Description         string           `json:"description"`
	Hardware            string           `json:"hardware"`
	Bandwidth           int              `json:"bandwidth"`
	MTU                 int              `json:"mtu"`
	InterfaceMembership string           `json:"interface_membership"`
	// Duplex              string           // `json:"duplex"`
	// AutoNegotiate       string           // `json:"auto_negotiate"`
	// Lanes               string           // `json:"lanes"`
}

type MACTableEntry struct {
	Ifname string           `json:"ifname"`
	Port   int              `json:"port"`
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

type InterfaceTable map[int]*InterfaceStatus

type MACTable map[string]*MACTableEntry

type LLDPNeighbors map[string]*LLDP

type SwitchHostTree struct {
	Host      *Host  `json:"host"`
	LLDP      *LLDP  `json:"lldp"`
	Interface string `json:"interface"`
}

type SwitchHostTreeList []SwitchHostTree

type SwitchLLDPList []LLDP

func (mt MACTable) Port(port int) []*MACTableEntry {
	entries := make([]*MACTableEntry, 0)
	for _, entry := range mt {
		if entry.Port == port {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (m *MACTableEntry) MarshalJSON() ([]byte, error) {
	type Alias MACTableEntry
	return json.Marshal(&struct {
		MAC string `json:"mac"`
		*Alias
	}{
		MAC:   m.MAC.String(),
		Alias: (*Alias)(m),
	})
}

func (mt MACTable) String() string {
	data, _ := json.MarshalIndent(mt, "", "    ")
	return string(data)
}
