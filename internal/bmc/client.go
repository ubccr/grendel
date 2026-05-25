// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"net"
	"net/http"

	"github.com/spf13/viper"
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
		HTTPClient: &http.Client{
			Timeout: viper.GetDuration("bmc.gofish.client_timeout"),
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   viper.GetDuration("bmc.gofish.dial_timeout"),
					KeepAlive: viper.GetDuration("bmc.gofish.dial_keep_alive_timeout"),
				}).DialContext,
				TLSHandshakeTimeout: viper.GetDuration("bmc.gofish.tls_handshake_timeout"),
				IdleConnTimeout:     viper.GetDuration("bmc.gofish.idle_conn_timeout"),
			},
		},
		MaxConcurrentRequests: viper.GetInt64("bmc.max_concurrent_request"),
		ReuseConnections:      viper.GetBool("bmc.reuse_connections"),
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
