package bmc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPMI(t *testing.T) {
	endpoint := os.Getenv("GRENDEL_IPMI_ENDPOINT")
	user := os.Getenv("GRENDEL_IPMI_USER")
	pass := os.Getenv("GRENDEL_IPMI_PASS")

	if endpoint == "" || user == "" || pass == "" {
		t.Skip("Skipping IPMI test. Missing env vars")
	}

	r, err := NewIPMI(endpoint, user, pass, 623)
	assert.Nil(t, err)

	system, err := r.GetSystem()
	assert.Nil(t, err)
	assert.Greater(t, len(system.BIOSVersion), 0)
}
