// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"fmt"
	"os"
	"testing"
)

func TestSonic(t *testing.T) {
	endpoint := os.Getenv("GRENDEL_SONIC_ENDPOINT")
	user := os.Getenv("GRENDEL_SONIC_USER")
	pass := os.Getenv("GRENDEL_SONIC_PASS")

	if endpoint == "" || user == "" || pass == "" {
		t.Skip("Skipping SONiC test. Missing env vars")
	}

	client, err := NewSonic(endpoint, user, pass, "", true)
	if err != nil {
		t.Fatal(err)
	}

	macTable, err := client.GetMACTable()
	if err != nil {
		t.Fatal(err)
	}

	if len(macTable) == 0 {
		t.Errorf("No mac table entries returned from api")
	}

	for _, entry := range macTable {
		fmt.Printf("%s - %d\n", entry.MAC, entry.Port)
	}
}
