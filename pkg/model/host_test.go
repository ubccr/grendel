// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
)

func TestHostTags(t *testing.T) {
	assert := assert.New(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	host.Tags = []string{"k11", "switch", "dellztd"}

	assert.True(host.HasTags("k11", "dellztd"))
	assert.False(host.HasTags("p22", "dellztd"))
	assert.True(host.HasAnyTags("p22", "dellztd"))
	assert.False(host.HasAnyTags("p22", "m12"))
	assert.False(host.HasAnyTags())
	assert.False(host.HasTags())
	assert.Equal("", host.Interfaces[0].HostNameIndex(100))
	assert.Equal(host.Interfaces[0].FQDN, host.Interfaces[0].HostNameIndex(0))
}

func TestHostBonds(t *testing.T) {
	assert := assert.New(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	assert.Equal(host.Bonds[0].AddrString(), host.Bonds[0].IP.Addr().String())
}
