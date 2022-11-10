// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"

	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/ubccr/grendel/firmware"
)

type Host struct {
	ID         ksuid.KSUID     `json:"id,omitempty"`
	Name       string          `json:"name" validate:"required,hostname"`
	Interfaces []*NetInterface `json:"interfaces"`
	Provision  bool            `json:"provision"`
	Firmware   firmware.Build  `json:"firmware"`
	BootImage  string          `json:"boot_image"`
	Tags       []string        `json:"tags"`
}

func (h *Host) HasTags(tags ...string) bool {
	for _, a := range tags {
		found := false
		for _, b := range h.Tags {
			if a == b {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return len(tags) > 0
}

func (h *Host) HasAnyTags(tags ...string) bool {
	for _, a := range h.Tags {
		for _, b := range tags {
			if a == b {
				return true
			}
		}
	}

	return false
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

func (h *Host) BootInterface() *NetInterface {
	for _, nic := range h.Interfaces {
		if !nic.BMC {
			return nic
		}
	}

	return nil
}

func (h *Host) FromJSON(hostJSON string) {
	h.Name = gjson.Get(hostJSON, "name").String()
	h.BootImage = gjson.Get(hostJSON, "boot_image").String()
	h.Provision = gjson.Get(hostJSON, "provision").Bool()
	h.ID, _ = ksuid.Parse(gjson.Get(hostJSON, "id").String())
	h.Firmware = firmware.NewFromString(gjson.Get(hostJSON, "firmware").String())

	h.Interfaces = make([]*NetInterface, 0)
	res := gjson.Get(hostJSON, "interfaces")
	for _, i := range res.Array() {
		nic := &NetInterface{}
		nic.Name = i.Get("ifname").String()
		nic.FQDN = i.Get("fqdn").String()
		nic.BMC = i.Get("bmc").Bool()
		nic.IP = net.ParseIP(i.Get("ip").String())
		nic.MAC, _ = net.ParseMAC(i.Get("mac").String())
		h.Interfaces = append(h.Interfaces, nic)
	}

	tres := gjson.Get(hostJSON, "tags")
	for _, i := range tres.Array() {
		h.Tags = append(h.Tags, i.String())
	}
}

func (h *Host) ToJSON() string {
	hostJSON := `{"firmware": "", "interfaces": [], "name": "", "provision": false, "kickstart": false, "boot_image": "", "tags": []}`

	if !h.ID.IsNil() {
		hostJSON, _ = sjson.Set(hostJSON, "id", h.ID.String())
	}
	hostJSON, _ = sjson.Set(hostJSON, "name", h.Name)
	hostJSON, _ = sjson.Set(hostJSON, "boot_image", h.BootImage)
	hostJSON, _ = sjson.Set(hostJSON, "firmware", h.Firmware.String())
	hostJSON, _ = sjson.Set(hostJSON, "provision", h.Provision)

	for _, nic := range h.Interfaces {
		n := map[string]interface{}{
			"mac":    nic.MAC.String(),
			"ip":     nic.IP.String(),
			"ifname": nic.Name,
			"fqdn":   nic.FQDN,
			"bmc":    nic.BMC,
		}
		hostJSON, _ = sjson.Set(hostJSON, "interfaces.-1", n)
	}

	for _, t := range h.Tags {
		hostJSON, _ = sjson.Set(hostJSON, "tags.-1", t)
	}

	return hostJSON
}

func (h *Host) MarshalJSON() ([]byte, error) {
	type Alias Host
	aux := &struct {
		ID       string `json:"id,omitempty"`
		Firmware string `json:"firmware"`
		*Alias
	}{
		Firmware: h.Firmware.String(),
		Alias:    (*Alias)(h),
	}

	if h.ID.IsNil() {
		aux.ID = ""
	} else {
		aux.ID = h.ID.String()
	}

	data, err := json.Marshal(&aux)
	if err != nil {
		return nil, err
	}

	return data, nil
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
