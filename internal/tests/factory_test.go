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
		assert.False(host.ID.IsNil())

		image := BootImageFactory.MustCreate().(*model.BootImage)
		assert.Greater(len(image.Name), 1)
	}
}
