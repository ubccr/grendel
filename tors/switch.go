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
