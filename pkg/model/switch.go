package model

import (
	"encoding/json"
	"net"
)

type InterfaceStatus struct {
	Name               string           `json:"name"`
	LineProtocolStatus string           `json:"line_protocol_status"`
	InterfaceStatus    string           `json:"interface_status"`
	PhysicalAddress    net.HardwareAddr `json:"physical_address"`
	Description        string           `json:"description"`
	Hardware           string           `json:"hardware"`
	// Bandwidth is in Mb/s.
	Bandwidth           int    `json:"bandwidth"`
	MTU                 int    `json:"mtu"`
	InterfaceMembership string `json:"interface_membership"`
	// Duplex              string           // `json:"duplex"`
	// AutoNegotiate       string           // `json:"auto_negotiate"`
	// Lanes               string           // `json:"lanes"`
}

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

type InterfaceTable map[string]*InterfaceStatus

type MACTable []*MACTableEntry

type LLDPNeighbors map[string]*LLDP

type SwitchLLDPList []LLDP

func (mt MACTable) ByIfname() map[string][]*MACTableEntry {
	out := make(map[string][]*MACTableEntry)
	for _, entry := range mt {
		out[entry.Ifname] = append(out[entry.Ifname], entry)
	}
	return out
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
