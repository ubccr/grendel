package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/firmware"
)

func TestToken(t *testing.T) {
	assert := assert.New(t)

	host := HostFactory.MustCreate().(*Host)
	token, err := NewFirmwareToken(host.Interfaces[0].MAC.String(), firmware.SNPONLY)
	if assert.NoError(err) {
		assert.Less(len(token), 128)
	}
}
