// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/settings/settingstest"
)

// the first subnet holding a host is the one used, so the order they come in decides between two that overlap
func TestSubnetsAreTriedInNameOrder(t *testing.T) {
	settingstest.Set(t, "dhcp.subnets", map[string]any{
		"10-everything": map[string]any{"gateway": "10.0.0.1/8"},
		"2-lab":         map[string]any{"gateway": "10.1.0.254/24", "dns": "10.1.0.2, 10.1.0.3", "mtu": 9000},
		"rack-f":        map[string]any{"gateway": "10.2.0.254/24"},
	})

	subnets := Subnets()
	require.Len(t, subnets, 3)
	assert.Equal(t, "10.1.0.254/24", subnets[0].Gateway.String(), "2-lab sorts before 10-everything")
	assert.Equal(t, "10.0.0.1/8", subnets[1].Gateway.String())
	assert.Equal(t, "10.2.0.254/24", subnets[2].Gateway.String())
	assert.Len(t, subnets[0].DNS, 2)
	assert.Equal(t, uint16(9000), subnets[0].MTU)
}

// a config file from before subnets had names lists them, and they keep the order they are listed in
func TestListedSubnetsKeepTheirOrder(t *testing.T) {
	rows := []map[string]any{}
	for _, gw := range []string{"10.9.0.1/24", "10.1.0.1/24", "10.5.0.1/24", "10.2.0.1/24", "10.8.0.1/24", "10.3.0.1/24", "10.7.0.1/24", "10.4.0.1/24", "10.6.0.1/24", "10.0.0.1/8", "10.10.0.1/24"} {
		rows = append(rows, map[string]any{"gateway": gw})
	}
	settingstest.Set(t, "dhcp.subnets", rows)

	subnets := Subnets()
	require.Len(t, subnets, len(rows))
	for i, row := range rows {
		assert.Equal(t, row["gateway"], subnets[i].Gateway.String())
	}
}
