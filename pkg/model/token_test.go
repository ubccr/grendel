// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/internal/firmware"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
)

func TestToken(t *testing.T) {
	assert := assert.New(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	token, err := model.NewFirmwareToken(host.Interfaces[0].MAC.String(), firmware.SNPONLYx86_64)
	if assert.NoError(err) {
		assert.Less(len(token), 128)
		assert.Greater(len(token), 0)
	}

	build, err := model.ParseFirmwareToken(token)
	if assert.NoError(err) {
		assert.Equal(build, firmware.SNPONLYx86_64)
	}

	token, err = model.NewBootToken(host.UID.String(), host.Interfaces[0].MAC.String())
	if assert.NoError(err) {
		assert.Greater(len(token), 0)
	}

	claims, err := model.ParseBootToken(token)
	if assert.NoError(err) {
		assert.Equal(claims.ID, host.UID.String())
		assert.Equal(claims.MAC, host.Interfaces[0].MAC.String())
	}
}
