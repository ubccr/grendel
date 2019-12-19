package model

import (
	"encoding/json"
	"fmt"
	"net"
)

type NetInterface struct {
	MAC  net.HardwareAddr `json:"mac" badgerhold:"unique" validate:"required"`
	Name string           `json:"ifname"`
	IP   net.IP           `json:"ip"`
	FQDN string           `json:"fqdn"`
	BMC  bool             `json:"bmc"`
}

func (n *NetInterface) MarshalJSON() ([]byte, error) {
	type Alias NetInterface
	return json.Marshal(&struct {
		MAC string `json:"mac"`
		IP  string `json:"ip"`
		*Alias
	}{
		MAC:   n.MAC.String(),
		IP:    n.IP.String(),
		Alias: (*Alias)(n),
	})
}

func (n *NetInterface) UnmarshalJSON(data []byte) error {
	type Alias NetInterface
	aux := &struct {
		MAC string `json:"mac"`
		IP  string `json:"ip"`
		*Alias
	}{
		Alias: (*Alias)(n),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	mac, err := net.ParseMAC(aux.MAC)
	if err != nil {
		return err
	}
	ip := net.ParseIP(aux.IP)
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("Invalid IPv4 address: %s", aux.IP)
	}

	n.MAC = mac
	n.IP = ip
	return nil
}
