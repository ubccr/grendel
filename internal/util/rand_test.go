// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRand(t *testing.T) {
	assert := assert.New(t)

	secret, err := GenerateSecret(32)
	if assert.NoError(err) {
		assert.Equal(len(secret), 64)
	}
}
