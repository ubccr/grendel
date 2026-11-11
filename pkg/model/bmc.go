// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import (
	"github.com/stmcginnis/gofish/oem/dell"
)

// JobMessage is one node's outcome as the event log records it
type JobMessage struct {
	Status string `json:"status"`
	Host   string `json:"host"`
	Msg    string `json:"msg"`
}

type JobMessageList []JobMessage

type RedfishSystem struct {
	Name           string  `json:"name"`
	HostName       string  `json:"host_name"`
	BIOSVersion    string  `json:"bios_version"`
	SerialNumber   string  `json:"serial_number"`
	Manufacturer   string  `json:"manufacturer"`
	Model          string  `json:"model"`
	PowerStatus    string  `json:"power_status"`
	Health         string  `json:"health"`
	TotalMemory    float32 `json:"total_memory"`
	ProcessorCount int     `json:"processor_count"`
	ProcessorModel string  `json:"processor_model"`
	// ProcessorSockets and ProcessorCores are zero when the bmc does not report them
	ProcessorSockets int `json:"processor_sockets"`
	ProcessorCores   int `json:"processor_cores"`
	// BootProgress is the last boot state the bmc saw, such as OSRunning, and BootProgressTime when it got there. Empty on firmware that does not report boot progress
	BootProgress       string `json:"boot_progress"`
	BootProgressTime   string `json:"boot_progress_time"`
	BootOverrideTarget string `json:"boot_override_target"`
	// BootOverrideEnabled is Disabled, Once or Continuous
	BootOverrideEnabled string          `json:"boot_override_enabled"`
	BootMode            string          `json:"boot_mode" description:"UEFI or Legacy"`
	BootNext            string          `json:"boot_next"`
	BootOrder           []string        `json:"boot_order" oai3:"nullable"`
	OEMDell             *dell.OEMSystem `json:"oem_dell" oai3:"nullable"`
}

// BmcSummary counts a run. FailedNodes is a nodeset string, so it can be pasted straight back into another command
type BmcSummary struct {
	Total       int    `json:"total"`
	Succeeded   int    `json:"succeeded"`
	Failed      int    `json:"failed"`
	FailedNodes string `json:"failed_nodes"`
	DurationMs  int64  `json:"duration_ms"`
}

// BmcResult is one node's outcome. Message says why a node failed. Data is present only on success, and only when the operation returns anything
type BmcResult[T any] struct {
	Node       string `json:"node"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	DurationMs int64  `json:"duration_ms"`
	Data       *T     `json:"data" oai3:"nullable"`
}

// BmcResponse is the envelope of every bmc route. RunID appears in the server log and the event record, which ties them to this response
type BmcResponse[T any] struct {
	RunID   string         `json:"run_id"`
	Summary BmcSummary     `json:"summary"`
	Results []BmcResult[T] `json:"results" oai3:"nullable"`
}

// BmcAck is the payload of an operation that reports nothing but whether it worked, so data is always null. What a node has to say, such as a power reset being substituted, arrives in the result's message
type BmcAck struct{}

// BmcJob is one job queued on a BMC. It is Grendel's own shape rather than gofish's schemas.Job, whose Messages hold slices with no omitempty. Those serialize as null, which the generated client rejects
type BmcJob struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	State           string `json:"state"`
	Status          string `json:"status"`
	Type            string `json:"type"`
	PercentComplete int    `json:"percent_complete"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	Message         string `json:"message"`
}

// BmcJobList is the payload of a jobs query. It is a struct rather than a bare slice so a node with an empty queue is distinguishable from one that was never reached
type BmcJobList struct {
	Jobs []BmcJob `json:"jobs" oai3:"nullable"`
}

// BmcSELEntry is one system event log record, narrowed the same way BmcJob is. gofish's schemas.LogEntry carries the diagnostic data blob and several slices with no omitempty
type BmcSELEntry struct {
	ID         string `json:"id"`
	Created    string `json:"created"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	MessageID  string `json:"message_id"`
	EntryType  string `json:"entry_type"`
	SensorType string `json:"sensor_type"`
	Resolution string `json:"resolution"`
}

// BmcSELLog is one node's system event log, newest entry first
type BmcSELLog struct {
	Entries []BmcSELEntry `json:"entries" oai3:"nullable"`
}

// BmcFirmwareUpdate is one package the iDRAC found newer in the catalog than what is installed
type BmcFirmwareUpdate struct {
	Name             string `json:"name"`
	DisplayName      string `json:"display_name"`
	ComponentType    string `json:"component_type"`
	InstalledVersion string `json:"installed_version"`
	PackageVersion   string `json:"package_version"`
	Criticality      string `json:"criticality" description:"1 = Warning, 2 = Critical, 3 = Recommended"`
	RebootType       string `json:"reboot_type" description:"NONE, IDRAC (non disruptive) or HOST (full power cycle)"`
	PackagePath      string `json:"package_path"`
}

// BmcFirmwareUpdates is one node's available updates. RebootType is the most disruptive reboot any of them needs
type BmcFirmwareUpdates struct {
	Updates    []BmcFirmwareUpdate `json:"updates" oai3:"nullable"`
	RebootType string              `json:"reboot_type" description:"NONE, IDRAC or HOST"`
}
