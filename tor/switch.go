package tor

import (
	"github.com/ubccr/grendel/logger"
)

var log = logger.GetLogger("SWITCH")

type MACTableEntry struct {
	Ifname string `json:"ifname"`
	Port   int    `json:"port"`
	VLAN   string `json:"vlan"`
	Type   string `json:"type"`
}

type MACTable map[string]*MACTableEntry

type NetworkSwitch interface {
	GetMACTable() (MACTable, error)
}
