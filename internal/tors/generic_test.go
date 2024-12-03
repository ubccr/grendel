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

package tors

import (
	"fmt"
	"os"
	"testing"
)

func TestGeneric(t *testing.T) {
	endpoint := os.Getenv("GRENDEL_SNMP_ENDPOINT")

	if endpoint == "" {
		t.Skip("Skipping generic snmp test. Missing env vars")
	}

	client, err := NewGeneric(endpoint, "public")
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
