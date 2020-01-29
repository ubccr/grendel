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
