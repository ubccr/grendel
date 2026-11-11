// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBiosDiff(t *testing.T) {
	nodes := map[string]map[string]string{
		"cpn-01": {"Sriov": "Enabled", "CStates": "Disabled", "NumLock": "On"},
		"cpn-02": {"Sriov": "Enabled", "CStates": "Disabled", "NumLock": "On"},
		"cpn-03": {"Sriov": "Disabled", "CStates": "Disabled", "NumLock": "On"},
		"cpn-04": {"Sriov": "Enabled", "CStates": "Disabled"},
	}

	diff := biosDiff(nodes, nil)
	require.Len(t, diff, 2, "CStates matches everywhere")

	assert.Equal(t, []biosValueGroup{
		{Value: "Enabled", Nodes: []string{"cpn-01", "cpn-02", "cpn-04"}},
		{Value: "Disabled", Nodes: []string{"cpn-03"}},
	}, diff["Sriov"], "the most common value comes first")
	assert.Equal(t, []biosValueGroup{
		{Value: "On", Nodes: []string{"cpn-01", "cpn-02", "cpn-03"}},
		{Value: "(absent)", Nodes: []string{"cpn-04"}},
	}, diff["NumLock"])

	assert.Empty(t, biosDiff(nodes, []string{"CStates"}), "naming attributes limits the comparison to them")
}

func TestParseAssignments(t *testing.T) {
	got, err := parseAssignments([]string{"Sriov=Enabled", "Port=3", "Banner=a=b"})
	require.NoError(t, err)
	assert.Equal(t, `"Enabled"`, string(got["Sriov"]))
	assert.Equal(t, `"3"`, string(got["Port"]), "sent as a string, the server converts it")
	assert.Equal(t, `"a=b"`, string(got["Banner"]), "only the first = splits")

	for _, bad := range [][]string{{"Sriov"}, {"=x"}, {"A=1", "A=2"}} {
		_, err := parseAssignments(bad)
		assert.Error(t, err, "%v", bad)
	}
}
