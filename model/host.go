package model

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/ubccr/grendel/firmware"
)

type NetworkAddress struct {
	MAC  net.HardwareAddr `json:"mac"`
	IP   net.IP           `json:"ip"`
	FQDN string           `json:"fqdn"`
}

type Host struct {
	MAC        net.HardwareAddr `json:"mac" badgerhold:"index" validate:"required"`
	IP         net.IP           `json:"ip" validate:"required"`
	FQDN       string           `json:"fqdn" validate:"required,fqdn"`
	Provision  bool             `json:"provision"`
	BMCAddress *NetworkAddress  `json:"bmc_address"`
	Firmware   firmware.Build
}

func (h *Host) MarshalJSON() ([]byte, error) {
	type Alias Host
	return json.Marshal(&struct {
		MAC string `json:"mac"`
		IP  string `json:"ip"`
		*Alias
	}{
		MAC:   h.MAC.String(),
		IP:    h.IP.String(),
		Alias: (*Alias)(h),
	})
}

func (h *Host) UnmarshalJSON(data []byte) error {
	type Alias Host
	aux := &struct {
		MAC string `json:"mac"`
		IP  string `json:"ip"`
		*Alias
	}{
		Alias: (*Alias)(h),
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

	h.MAC = mac
	h.IP = ip
	return nil
}

func (h *NetworkAddress) MarshalJSON() ([]byte, error) {
	type Alias NetworkAddress
	return json.Marshal(&struct {
		MAC string `json:"mac"`
		IP  string `json:"ip"`
		*Alias
	}{
		MAC:   h.MAC.String(),
		IP:    h.IP.String(),
		Alias: (*Alias)(h),
	})
}

func (h *NetworkAddress) UnmarshalJSON(data []byte) error {
	type Alias NetworkAddress
	aux := &struct {
		MAC string `json:"mac"`
		IP  string `json:"ip"`
		*Alias
	}{
		Alias: (*Alias)(h),
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

	h.MAC = mac
	h.IP = ip
	return nil
}
