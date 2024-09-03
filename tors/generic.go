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

package tors

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/alouca/gosnmp"
)

const (
	dot1qTpFdbAddress = ".1.3.6.1.2.1.17.7.1.2.2.1.2."
)

type Generic struct {
	endpoint  string
	community string
}

func NewGeneric(endpoint, community string) (*Generic, error) {
	g := &Generic{endpoint: endpoint, community: community}

	return g, nil
}

func (g *Generic) GetMACTable() (MACTable, error) {
	macTable := make(MACTable, 0)

	client, err := gosnmp.NewGoSNMP(g.endpoint, g.community, gosnmp.Version2c, 15)
	if err != nil {
		return nil, err
	}

	results, err := client.Walk(dot1qTpFdbAddress)
	for _, rec := range results {
		if rec.Type != gosnmp.Integer {
			log.Warnf("Invalid result type. Expecting Integer got: %d", rec.Type)
			continue
		}

		key := strings.Split(strings.TrimPrefix(rec.Name, dot1qTpFdbAddress), ".")
		idx := 1
		switch len(key) {
		case 7:
			idx = 1
		case 8:
			idx = 2
		default:
			log.Warnf("Invalid oid string: %s", rec.Name)
			continue
		}

		format := make([]string, 0, len(key)-idx)
		for _, i := range key[idx:] {
			val, _ := strconv.Atoi(i)
			format = append(format, fmt.Sprintf("%02x", val))
		}

		macStr := strings.Join(format, ":")
		mac, err := net.ParseMAC(macStr)
		if err != nil {
			log.Errorf("Invalid mac address %s: %v", macStr, err)
			continue
		}

		macTable[macStr] = &MACTableEntry{
			Port: rec.Value.(int) - 1,
			VLAN: key[0],
			MAC:  mac,
		}
	}

	log.Infof("Received %d entries", len(macTable))
	return macTable, nil
}

func (g *Generic) GetLLDPNeighbors() (LLDPNeighbors, error) {
	return nil, errors.New("LLDP not supported with Generic switch type")
}
func (g *Generic) GetInterfaceStatus() (InterfaceTable, error) {
	return nil, errors.New("Interface status not supported with Generic switch type")
}
