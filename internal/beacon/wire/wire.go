// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package wire is the report the beacon posts to /boot/:token/status. The beacon and the provision server both import it, and it imports nothing of grendel's, so the beacon binary stays small
package wire

import "regexp"

// States a report can carry, which are the provision states a node reports for itself
const (
	StateBooting   = "booting"
	StateFailed    = "failed"
	StateCompleted = "completed"
)

type FailedUnit struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ActiveState string `json:"active_state,omitempty"`
	SubState    string `json:"sub_state,omitempty"`
	Result      string `json:"result,omitempty"`
}

type Report struct {
	State       string       `json:"state"`
	Message     string       `json:"message,omitempty"`
	FailedUnits []FailedUnit `json:"failed_units,omitempty"`
	QueuedJobs  int          `json:"queued_jobs,omitempty"`
	BootID      string       `json:"boot_id,omitempty"`
	// Complete also turns provisioning off for the node, as the complete endpoint does, when State is completed
	Complete bool `json:"complete,omitempty"`
	// Facts are merged into the ones already reported for this provision, a later value for a key replacing the earlier one
	Facts map[string]string `json:"facts,omitempty"`
}

var factKey = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{0,63}$`)

// ValidFactKey reports whether key is a dotted lowercase name such as os.image_version or slurm.version, which is all the server keeps
func ValidFactKey(key string) bool {
	return factKey.MatchString(key)
}
