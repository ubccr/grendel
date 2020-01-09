package bmc

import (
	"os"
	"testing"
)

func TestRedfish(t *testing.T) {
	endpoint := os.Getenv("GRENDEL_BMC_ENDPOINT")
	user := os.Getenv("GRENDEL_BMC_USER")
	pass := os.Getenv("GRENDEL_BMC_PASS")

	if endpoint == "" || user == "" || pass == "" {
		t.Skip("Skipping BMC test. Missing env vars")
	}

	r, err := NewRedfish(endpoint, user, pass, true)
	if err != nil {
		t.Fatal(err)
	}

	service := r.client.Service
	_, err = service.Systems()
	if err != nil {
		t.Fatal(err)
	}
}
