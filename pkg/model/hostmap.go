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
	"sync"
)

type HostMap struct {
	sync.RWMutex
	internal map[string]*Host
}

func NewHostMap() *HostMap {
	return &HostMap{
		internal: make(map[string]*Host),
	}
}

func (hm *HostMap) Load(key string) (*Host, bool) {
	hm.RLock()
	result, ok := hm.internal[key]
	hm.RUnlock()
	return result, ok
}

func (hm *HostMap) Delete(key string) {
	hm.Lock()
	delete(hm.internal, key)
	hm.Unlock()
}

func (hm *HostMap) Store(key string, value *Host) {
	hm.Lock()
	hm.internal[key] = value
	hm.Unlock()
}
