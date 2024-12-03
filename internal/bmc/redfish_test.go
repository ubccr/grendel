// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedfish(t *testing.T) {
	endpoint := os.Getenv("GRENDEL_BMC_ENDPOINT")
	user := os.Getenv("GRENDEL_BMC_USER")
	pass := os.Getenv("GRENDEL_BMC_PASS")

	if endpoint == "" || user == "" || pass == "" {
		t.Skip("Skipping BMC test. Missing env vars")
	}

	r, err := NewRedfishClient(endpoint, user, pass, true)
	assert.Nil(t, err)
	defer r.client.Logout()

	system, err := r.GetSystem()
	assert.Nil(t, err)
	assert.Greater(t, len(system.BIOSVersion), 0)
}
