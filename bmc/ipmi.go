package bmc

import (
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
