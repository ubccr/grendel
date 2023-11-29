package frontend

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/ubccr/grendel/nodeset"
	"github.com/ubccr/grendel/tors"
)

func (h *Handler) getMacAddress(switchName string) (tors.MACTable, error) {
	nodeset, err := nodeset.NewNodeSet(switchName)
	if err != nil {
		return nil, err
	}
	host, err := h.DB.FindHosts(nodeset)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("https://%s", host[0].InterfaceBMC().ToStdAddr().String())
	sw, err := tors.NewDellOS10(endpoint, "admin", viper.GetString("bmc.switch_admin_password"), "", true)
	if err != nil {
		return nil, err
	}

	macTable, err := sw.GetMACTable()
	if err != nil {
		return nil, err
	}
	return macTable, nil
}
