// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"fmt"
	"os"
	"testing"
)

func TestArista(t *testing.T) {
	endpoint := os.Getenv("GRENDEL_ARISTA_ENDPOINT")
	user := os.Getenv("GRENDEL_ARISTA_USER")
	pass := os.Getenv("GRENDEL_ARISTA_PASS")

	if endpoint == "" || user == "" || pass == "" {
		t.Skip("Skipping Arista test. Missing env vars")
	}

	client, err := NewArista(endpoint, user, pass)
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

	lldpNeighbors, err := client.GetLLDPNeighbors()
	if err != nil {
		t.Fatal(err)
	}

	if len(lldpNeighbors) == 0 {
		t.Errorf("No lldp neighbors returned from api")
	}

	for iface, entry := range lldpNeighbors {
		fmt.Printf("%s - %s - %s\n", iface, entry.ChassisId, entry.PortDescription)
	}
}
