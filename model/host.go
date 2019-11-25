package model

import (
	"net"
)

type Host struct {
	MAC      net.HardwareAddr
	IP       net.IP
	FQDN     string
	BootSpec string
}
