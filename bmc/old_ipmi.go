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
	"fmt"

	"github.com/vmware/goipmi"
)

type IPMI struct {
	client *ipmi.Client
}

func NewIPMI(hostname, user, pass string, port int) (*IPMI, error) {
	conn := &ipmi.Connection{
		Interface: "lanplus",
		Hostname:  hostname,
		Port:      port,
		Username:  user,
		Password:  pass,
	}
	client, err := ipmi.NewClient(conn)
	if err != nil {
		return nil, err
	}

	return &IPMI{client: client}, nil
}

func (i *IPMI) Logout() {
	i.client.Close()
}

func (i *IPMI) powerControl(ctl ipmi.ChassisControl) error {
	err := i.client.Open()
	if err != nil {
		return err
	}

	err = i.client.Control(ctl)
	if err != nil {
		return err
	}

	err = i.client.Close()
	if err != nil {
		return err
	}

	return nil

}

func (i *IPMI) PowerCycle() error {
	return i.powerControl(ipmi.ControlPowerCycle)
}

func (i *IPMI) PowerOn() error {
	return i.powerControl(ipmi.ControlPowerUp)
}

func (i *IPMI) PowerOff() error {
	return i.powerControl(ipmi.ControlPowerDown)
}

func (i *IPMI) EnablePXE() error {
	err := i.client.Open()
	if err != nil {
		return err
	}

	err = i.client.SetBootDevice(ipmi.BootDevicePxe)
	if err != nil {
		return err
	}

	err = i.client.Close()
	if err != nil {
		return err
	}

	return nil
}

func (i *IPMI) GetSystem() (*System, error) {
	err := i.client.Open()
	if err != nil {
		return nil, err
	}

	r := &ipmi.Request{
		ipmi.NetworkFunctionChassis,
		ipmi.CommandChassisStatus,
		&ipmi.ChassisStatusRequest{},
	}

	status := &ipmi.ChassisStatusResponse{}
	err = i.client.Send(r, status)
	if err != nil {
		return nil, err
	}

	deviceID, err := i.client.DeviceID()
	if err != nil {
		return nil, err
	}

	system := &System{
		BIOSVersion:  fmt.Sprintf("%d.%d", deviceID.FirmwareRevision1, deviceID.FirmwareRevision2),
		PowerStatus:  status.String(),
		Manufacturer: deviceID.ManufacturerID.String(),
	}

	return system, nil
}
