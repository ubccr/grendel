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
	"encoding/json"
	"fmt"
	"net"
)

type NetInterface struct {
	MAC  net.HardwareAddr `json:"mac" validate:"required"`
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
	if aux.MAC != "" {
		mac, err := net.ParseMAC(aux.MAC)
		if err != nil {
			return err
		}

		n.MAC = mac
	}

	if aux.IP != "" {
		ip := net.ParseIP(aux.IP)
		if ip == nil || ip.To4() == nil {
			return fmt.Errorf("Invalid IPv4 address: %s", aux.IP)
		}

		n.IP = ip
	}
	return nil
}
