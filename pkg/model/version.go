// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

type Version struct {
	Version string
	// Commit and CommitTime are the git revision the server was built from and when that revision was committed. Both are empty for a build made outside a git checkout
	Commit     string `json:"commit,omitempty"`
	CommitTime string `json:"commit_time,omitempty"`
	// Modified is set when the checkout had uncommitted changes at build time
	Modified  bool   `json:"modified,omitempty"`
	GoVersion string `json:"go_version,omitempty"`
	// Schema is the database schema version this build migrates to
	Schema uint `json:"schema,omitempty"`
}
