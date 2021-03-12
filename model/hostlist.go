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
	"strings"

	"github.com/ubccr/grendel/nodeset"
)

type HostList []*Host

func (hl HostList) FilterPrefix(prefix string) HostList {
	n := 0
	for _, host := range hl {
		if strings.HasPrefix(host.Name, prefix) {
			hl[n] = host
			n++
		}
	}

	return hl[:n]
}

func (hl HostList) ToNodeSet() (*nodeset.NodeSet, error) {
	nodes := []string{}
	for _, host := range hl {
		nodes = append(nodes, host.Name)
	}

	return nodeset.NewNodeSet(strings.Join(nodes, ","))
}

func NewHostList() HostList {
	return make(HostList, 0)
}
