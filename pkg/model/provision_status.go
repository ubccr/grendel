// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import "time"

// ProvisionState is how far a node has got through provisioning
type ProvisionState string

const (
	// ProvisionStatePXE is set when the node's firmware asks DHCP for a network boot
	ProvisionStatePXE ProvisionState = "pxe"
	// ProvisionStateIPXE is set when the node fetches its iPXE script
	ProvisionStateIPXE ProvisionState = "ipxe"
	// ProvisionStateConfig is set when the booted OS fetches its ignition, kickstart or cloud-init config
	ProvisionStateConfig ProvisionState = "config"
	// ProvisionStateBooting is reported by the beacon while systemd is still starting units
	ProvisionStateBooting ProvisionState = "booting"
	// ProvisionStateFailed is reported by the beacon when the boot ended with failed units, timed out, or a script said so
	ProvisionStateFailed ProvisionState = "failed"
	// ProvisionStateCompleted is reported by the beacon on a clean boot, or set by the complete endpoint
	ProvisionStateCompleted ProvisionState = "completed"
)

// ProvisionStates lists every state, in the order a node goes through them
var ProvisionStates = []ProvisionState{
	ProvisionStatePXE,
	ProvisionStateIPXE,
	ProvisionStateConfig,
	ProvisionStateBooting,
	ProvisionStateFailed,
	ProvisionStateCompleted,
}

// Valid reports whether s is one of the known states
func (s ProvisionState) Valid() bool {
	for _, state := range ProvisionStates {
		if s == state {
			return true
		}
	}
	return false
}

// Finished reports whether the provision has ended, well or badly
func (s ProvisionState) Finished() bool {
	return s == ProvisionStateFailed || s == ProvisionStateCompleted
}

// FailedUnit is a systemd unit that failed during provisioning
type FailedUnit struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ActiveState string `json:"active_state,omitempty"`
	SubState    string `json:"sub_state,omitempty"`
	Result      string `json:"result,omitempty"`
}

// ProvisionStatus is the current state of a node's latest provision. Only the latest is kept, a new network boot starts it over
type ProvisionStatus struct {
	Node        string         `json:"node"`
	State       ProvisionState `json:"state"`
	StartedAt   time.Time      `json:"started_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	ProvisionStatusDetail
}

// ProvisionStatusDetail is the part of a status the beacon fills in, stored together as JSON
type ProvisionStatusDetail struct {
	Message     string       `json:"message,omitempty"`
	FailedUnits []FailedUnit `json:"failed_units,omitempty"`
	QueuedJobs  int          `json:"queued_jobs,omitempty"`
	BootID      string       `json:"boot_id,omitempty"`
	Addr        string       `json:"addr,omitempty"`
	// Stages is when the node first reached each state in this provision
	Stages map[ProvisionState]time.Time `json:"stages,omitempty"`
	// BootImage and BootImageVersion are the image the node was sent when it fetched its iPXE script, so what it reports running can be checked against it
	BootImage        string `json:"boot_image,omitempty"`
	BootImageVersion string `json:"boot_image_version,omitempty"`
	// Facts are what the beacon found on the node, such as os.image_version, keyed by dotted lowercase names
	Facts map[string]string `json:"facts,omitempty"`
}

// FactImageVersion is the fact holding IMAGE_VERSION from the node's os-release, compared against BootImageVersion
const FactImageVersion = "os.image_version"

type ProvisionStatusList []ProvisionStatus
