package bmc

import (
	"errors"
	"fmt"

	"github.com/ubccr/grendel/bmc"
	"github.com/ubccr/grendel/model"
)

func systemMgr(host *model.Host) (bmc.SystemManager, error) {
	bmcIntf := host.InterfaceBMC()
	if bmcIntf == nil {
		return nil, errors.New("BMC interface not found")
	}

	bmcAddress := bmcIntf.FQDN
	if bmcAddress == "" {
		bmcAddress = bmcIntf.IP.String()
	}

	if bmcAddress == "" {
		return nil, errors.New("BMC address not set")
	}

	if useIPMI {
		ipmi, err := bmc.NewIPMI(bmcAddress, bmcUser, bmcPassword, 623)
		if err != nil {
			return nil, err
		}

		return ipmi, nil
	}

	redfish, err := bmc.NewRedfish(fmt.Sprintf("https://%s", bmcAddress), bmcUser, bmcPassword, true)
	if err != nil {
		return nil, err
	}

	return redfish, nil
}
