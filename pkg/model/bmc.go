// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"github.com/stmcginnis/gofish/oem/dell"
	"github.com/stmcginnis/gofish/redfish"
)

type OSPower struct {
	Hosts       string                           `json:"hosts"`
	PowerOption redfish.ResetType                `json:"power_option"`
	BootOption  redfish.BootSourceOverrideTarget `json:"boot_option"`
}

type JobMessageList []JobMessage

type JobMessage struct {
	Status       string       `json:"status"`
	Host         string       `json:"host"`
	Msg          string       `json:"msg"`
	RedfishError RedfishError `json:"redfish_error"`
	Data         string       `json:"data"`
}

type RedfishJobList []RedfishJob
type RedfishJob struct {
	Host string         `json:"name"`
	Jobs []*redfish.Job `json:"jobs" oai3:"nullable"`
}

type RedfishMetricReportList []RedfishMetricReport
type RedfishMetricReport struct {
	Name    string                  `json:"name"`
	Reports []*redfish.MetricReport `json:"reports"`
}

type RedfishSystemList []RedfishSystem
type RedfishSystem struct {
	Name           string         `json:"name"`
	HostName       string         `json:"host_name"`
	BIOSVersion    string         `json:"bios_version"`
	SerialNumber   string         `json:"serial_number"`
	Manufacturer   string         `json:"manufacturer"`
	Model          string         `json:"model"`
	PowerStatus    string         `json:"power_status"`
	Health         string         `json:"health"`
	TotalMemory    float32        `json:"total_memory"`
	ProcessorCount int            `json:"processor_count"`
	BootNext       string         `json:"boot_next"`
	BootOrder      []string       `json:"boot_order" oai3:"nullable"`
	OEMDell        dell.OEMSystem `json:"oem_dell" oai3:"nullable"`
}

type RedfishDellUpgradeFirmwareList []RedfishDellUpgradeFirmware
type RedfishDellUpgradeFirmware struct {
	Name             string
	Status           string
	Message          string
	UpdateCount      int
	UpdateRebootType string
	UpdateList       dell.UpdateList //`oai3:"nullable"`
}

// TODO: verify correct json parsing
type RedfishError struct {
	Code  string `json:"code"`
	Error struct {
		MessageExtendedInfo []struct {
			Message                string `json:"Message"`
			MessageArgsCount       int    `json:"MessageArgs.@odata.count"`
			MessageId              string `json:"MessageId"`
			RelatedPropertiesCount int    `json:"RelatedProperties.@odata.count"`
			Resolution             string `json:"Resolution"`
			Severity               string `json:"Severity"`
		} `json:"@Message.ExtendedInfo,omitempty"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
