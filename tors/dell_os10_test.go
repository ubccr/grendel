package tors

import (
	//"fmt"
	"os"
	"testing"
)

func TestDellOS10(t *testing.T) {
	endpoint := os.Getenv("GRENDEL_DELLOS10_ENDPOINT")
	user := os.Getenv("GRENDEL_DELLOS10_USER")
	pass := os.Getenv("GRENDEL_DELLOS10_PASS")

	if endpoint == "" || user == "" || pass == "" {
		t.Skip("Skipping DellOS10 test. Missing env vars")
	}

	client, err := NewDellOS10(endpoint, user, pass, "", true)
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

	//	for _, entry := range macTable {
	//	fmt.Printf("%s - %d\n", entry.MAC, entry.Port)
	//	}
}
