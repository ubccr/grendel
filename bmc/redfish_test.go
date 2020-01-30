// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

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

	r, err := NewRedfish(endpoint, user, pass, true)
	assert.Nil(t, err)
	defer r.Logout()

	system, err := r.GetSystem()
	assert.Nil(t, err)
	assert.Greater(t, len(system.BIOSVersion), 0)
}
