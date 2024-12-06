// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/pkg/model"
)

func TestFactory(t *testing.T) {
	assert := assert.New(t)

	for i := 0; i < 3; i++ {
		host := HostFactory.MustCreate().(*model.Host)
		assert.Greater(len(host.Name), 1)
		assert.Equal(2, len(host.Interfaces))
		assert.Equal(1, len(host.Bonds))
		assert.False(host.UID.IsNil())

		image := BootImageFactory.MustCreate().(*model.BootImage)
		assert.Greater(len(image.Name), 1)
	}
}
