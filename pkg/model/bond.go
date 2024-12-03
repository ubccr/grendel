// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import "encoding/json"

type Bond struct {
	NetInterface
	Peers []string `json:"peers"`
}

func (b *Bond) MarshalJSON() ([]byte, error) {
	type Alias NetInterface
	return json.Marshal(&struct {
		Peers *[]string `json:"peers"`
		*Alias
	}{
		Peers: &b.Peers,
		Alias: (*Alias)(&b.NetInterface),
	})
}

func (b *Bond) UnmarshalJSON(data []byte) error {
	type Alias NetInterface
	aux := &struct {
		Peers *[]string `json:"peers"`
		*Alias
	}{
		Peers: &b.Peers,
		Alias: (*Alias)(&b.NetInterface),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}
