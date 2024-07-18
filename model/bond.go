// Copyright 2019 Grendel Authors. All rights reserved.  //
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

	"github.com/segmentio/ksuid"
)

type Bond struct {
	HostID ksuid.KSUID
	NetInterface
	Peers []string `json:"peers" gorm:"serializer:json"`
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
