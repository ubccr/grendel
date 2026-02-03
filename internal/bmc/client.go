// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/schemas"
)

type Redfish struct {
	config  gofish.ClientConfig
	client  *gofish.APIClient
	service *gofish.Service
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
	Jobs map[string]*schemas.Job
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
