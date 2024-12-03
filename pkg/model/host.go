// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/ubccr/grendel/internal/firmware"
)

type Host struct {
	ID         ksuid.KSUID     `json:"id,omitempty"`
	Name       string          `json:"name" validate:"required,hostname"`
	Interfaces []*NetInterface `json:"interfaces"`
	Bonds      []*Bond         `json:"bonds"`
	Provision  bool            `json:"provision"`
	Firmware   firmware.Build  `json:"firmware"`
	BootImage  string          `json:"boot_image"`
	Tags       []string        `json:"tags"`
}

func (h *Host) HostType() string {
	n := strings.Split(h.Name, "-")
	t := "server"
	if n[0] == "srv" || n[0] == "cpn" {
		t = "server"
	} else if n[0] == "swe" || n[0] == "swi" {
		t = "switch"
	} else if n[0] == "pdu" || n[0] == "ups" {
		t = "power"
	}
	return t
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

func (h *Host) InterfaceBonded(peer string) bool {
	for _, bond := range h.Bonds {
		for _, p := range bond.Peers {
			if peer == p {
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
		nic.VLAN = i.Get("vlan").String()
		nic.MTU = uint16(i.Get("mtu").Int())
		nic.IP, _ = netip.ParsePrefix(i.Get("ip").String())
		nic.MAC, _ = net.ParseMAC(i.Get("mac").String())
		h.Interfaces = append(h.Interfaces, nic)
	}

	h.Bonds = make([]*Bond, 0)
	bres := gjson.Get(hostJSON, "bonds")
	for _, i := range bres.Array() {
		bond := &Bond{Peers: []string{}}
		bond.Name = i.Get("ifname").String()
		bond.FQDN = i.Get("fqdn").String()
		bond.BMC = i.Get("bmc").Bool()
		bond.VLAN = i.Get("vlan").String()
		bond.MTU = uint16(i.Get("mtu").Int())
		bond.IP, _ = netip.ParsePrefix(i.Get("ip").String())
		bond.MAC, _ = net.ParseMAC(i.Get("mac").String())
		for _, p := range i.Get("peers").Array() {
			bond.Peers = append(bond.Peers, p.String())
		}
		h.Bonds = append(h.Bonds, bond)
	}

	tres := gjson.Get(hostJSON, "tags")
	for _, i := range tres.Array() {
		h.Tags = append(h.Tags, i.String())
	}
}

func (h *Host) ToJSON() string {
	hostJSON := `{"firmware": "", "interfaces": [], "bonds": [], "name": "", "provision": false, "kickstart": false, "boot_image": "", "tags": []}`

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
			"ip":     nic.CIDR(),
			"ifname": nic.Name,
			"fqdn":   nic.FQDN,
			"bmc":    nic.BMC,
			"vlan":   nic.VLAN,
			"mtu":    nic.MTU,
		}
		hostJSON, _ = sjson.Set(hostJSON, "interfaces.-1", n)
	}

	for _, bond := range h.Bonds {
		b := map[string]interface{}{
			"peers":  bond.Peers,
			"mac":    bond.MAC.String(),
			"ip":     bond.CIDR(),
			"ifname": bond.Name,
			"fqdn":   bond.FQDN,
			"bmc":    bond.BMC,
			"vlan":   bond.VLAN,
			"mtu":    bond.MTU,
		}
		hostJSON, _ = sjson.Set(hostJSON, "bonds.-1", b)
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
