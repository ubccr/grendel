// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ubccr/grendel/pkg/nodeset"
)

type HostList []*Host

func (hl *HostList) Scan(value interface{}) error {
	data, ok := value.(string)
	if !ok {
		return errors.New("incompatible type")
	}
	var hlist HostList
	err := json.Unmarshal([]byte(data), &hlist)
	if err != nil {
		return fmt.Errorf("failed to decode: %w", err)
	}

	*hl = hlist
	return nil
}

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
