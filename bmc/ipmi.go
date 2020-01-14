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

func (i *IPMI) PowerCycle() error {
	err := i.client.Open()
	if err != nil {
		return err
	}

	err = i.client.Control(ipmi.ControlPowerCycle)
	if err != nil {
		return err
	}

	err = i.client.Close()
	if err != nil {
		return err
	}

	return nil
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
