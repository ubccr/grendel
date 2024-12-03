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

	token, err = model.NewBootToken(host.ID.String(), host.Interfaces[0].MAC.String())
	if assert.NoError(err) {
		assert.Greater(len(token), 0)
	}

	claims, err := model.ParseBootToken(token)
	if assert.NoError(err) {
		assert.Equal(claims.ID, host.ID.String())
		assert.Equal(claims.MAC, host.Interfaces[0].MAC.String())
	}
}
