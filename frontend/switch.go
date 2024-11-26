package frontend

import (
	"errors"

	"github.com/ubccr/grendel/nodeset"
	"github.com/ubccr/grendel/tors"
)

func (h *Handler) getMacAddress(switchName string) (tors.MACTable, error) {
	nodeset, err := nodeset.NewNodeSet(switchName)
	if err != nil {
		return nil, err
	}
	hosts, err := h.DB.FindHosts(nodeset)
	if err != nil {
		return nil, err
	}
	if len(hosts) != 1 {
		return nil, errors.New("failed to load switch from DB")
	}
	host := hosts[0]

	sw, err := tors.NewNetworkSwitch(host)
	if err != nil {
		return nil, err
	}
	macTable, err := sw.GetMACTable()
	if err != nil {
		return nil, err
	}
	return macTable, nil
}
