package model

import (
	"net"
)

type Host struct {
	MAC      net.HardwareAddr `json:"mac" badgerhold:"index"`
	IP       net.IP           `json:"ip"`
	FQDN     string           `json:"fqdn"`
	BootSpec string           `json:bootspec"`
}
