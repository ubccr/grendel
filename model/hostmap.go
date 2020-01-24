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
