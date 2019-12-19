package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"

	"github.com/segmentio/ksuid"
	"github.com/ubccr/grendel/firmware"
)

type Host struct {
	ID         ksuid.KSUID     `json:"id" badgerhold:"index"`
	Name       string          `json:"name" badgerhold:"unique" validate:"required,hostname"`
	Interfaces []*NetInterface `json:"interfaces"`
	Provision  bool            `json:"provision"`
	Firmware   firmware.Build  `json:"firmware"`
}

func (h *Host) Interface(mac net.HardwareAddr) *NetInterface {
	for _, nic := range h.Interfaces {
		if bytes.Compare(nic.MAC, mac) == 0 {
			return nic
		}
	}

	return nil
}

func (h *Host) InterfaceBMC() *NetInterface {
	for _, nic := range h.Interfaces {
		if nic.BMC {
			return nic
		}
	}

	return nil
}

func (h *Host) MarshalJSON() ([]byte, error) {
	type Alias Host
	return json.Marshal(&struct {
		Firmware string `json:"firmware"`
		*Alias
	}{
		Firmware: h.Firmware.String(),
		Alias:    (*Alias)(h),
	})
}

func (h *Host) UnmarshalJSON(data []byte) error {
	type Alias Host
	aux := &struct {
		Firmware string `json:"firmware"`
		*Alias
	}{
		Alias: (*Alias)(h),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	h.Firmware = firmware.NewFromString(aux.Firmware)
	if len(aux.Firmware) != 0 && h.Firmware.IsNil() {
		return fmt.Errorf("Invalid firmware build: %s", aux.Firmware)
	}

	return nil
}
