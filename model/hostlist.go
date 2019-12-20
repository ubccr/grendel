package model

import (
	"strings"
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

func NewHostList() HostList {
	return make(HostList, 0)
}
