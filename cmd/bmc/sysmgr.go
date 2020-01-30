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
