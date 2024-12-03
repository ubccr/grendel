// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

type Redfish struct {
	config  gofish.ClientConfig
	client  *gofish.APIClient
	service *gofish.Service
}

type System struct {
	Name           string   `json:"name"`
	HostName       string   `json:"host_name"`
	BIOSVersion    string   `json:"bios_version"`
	SerialNumber   string   `json:"serial_number"`
	Manufacturer   string   `json:"manufacturer"`
	Model          string   `json:"model"`
	PowerStatus    string   `json:"power_status"`
	Health         string   `json:"health"`
	TotalMemory    float32  `json:"total_memory"`
	ProcessorCount int      `json:"processor_count"`
	BootNext       string   `json:"boot_next"`
	BootOrder      []string `json:"boot_order"`
	OEM            SystemOEM
}

type SystemOEM struct {
	Dell struct {
		DellSystem struct {
			ManagedSystemSize string
			MaxCPUSockets     int
			MaxDIMMSlots      int
			MacPCIeSlots      int
			SystemID          int
		}
	}
}

type Firmware struct {
	Name             string                     `json:"name"`
	SystemID         string                     `json:"system_id"`
	CurrentFirmwares map[string]CurrentFirmware `json:"current_firmware"`
}
type CurrentFirmware struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ReleaseDate string `json:"release_date"`
	SoftwareID  string `json:"software_id"`
	Updatable   bool   `json:"updateable"`
	Version     string `json:"version"`
}

type FirmwareUpdate struct {
	Firmware
	Jobs map[string]*redfish.Job
}

type BMCJob struct {
	Host string `json:"name"`
	Jobs []*redfish.Job
}

func NewRedfishClient(ip, user, pass string, insecure bool) (*Redfish, error) {
	endpoint := "https://" + ip

	config := gofish.ClientConfig{
		Endpoint: endpoint,
		Username: user,
		Password: pass,
		Insecure: insecure,
	}

	client, err := gofish.Connect(config)
	if err != nil {
		e := ParseRedfishError(err)
		// Try with default credentials
		if e.Code == "401" {
			config.Username = "root"
			config.Password = "calvin"
			client, err = gofish.Connect(config)
			if err != nil {
				log.Debug("default credentials failed")
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &Redfish{
		config:  config,
		client:  client,
		service: client.GetService(),
	}, nil
}
