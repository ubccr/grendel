// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"strings"

	"github.com/ubccr/grendel/pkg/nodeset"
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
	ns, err := nodeset.NewNodeSet("")
	if err != nil {
		return nil, err
	}

	for _, host := range hl {
		err := ns.Add(host.Name)
		if err != nil {
			return nil, err
		}
	}

	return ns, nil
}

func NewHostList() HostList {
	return make(HostList, 0)
}
