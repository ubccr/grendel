// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/pkg/model"
)

func render(t *testing.T, text string, host *model.Host) (string, error) {
	t.Helper()
	tmpl, err := parseIdentityTemplate("test", text)
	require.NoError(t, err)
	return renderIdentity(tmpl, host, position{4, 40})
}

func TestRenderIdentity(t *testing.T) {
	host := &model.Host{Name: "cpn-f01-04", Tags: []string{"gpu", "rack=f01", "dc=VTC"}}

	cases := []struct{ text, want string }{
		{"plain", "plain"},
		{"UB-{{.Name}}", "UB-cpn-f01-04"},
		{"{{ index .Parts 1 }}", "f01"},
		{"{{ index .Parts 2 | atoi }}", "4"},
		{`{{ tag "rack" | upper }}`, "F01"},
		{`{{ .Tags.dc }}`, "VTC"},
		{`{{ index .Parts 2 | trimLeft "0" }}`, "4"},
		{`{{ index .Tags "missing" | default "none" }}`, "none"},
		{"{{ .Index }} of {{ .Count }}", "4 of 40"},
		{`n{{ add .Index 1 | printf "%03d" }}`, "n005"},
	}
	for _, tc := range cases {
		got, err := render(t, tc.text, host)
		require.NoError(t, err, tc.text)
		assert.Equal(t, tc.want, got, tc.text)
	}

	_, err := render(t, "{{ index .Parts 5 }}", host)
	assert.Error(t, err, "an index past the end of the name fails the node")

	_, err = render(t, "{{ .Tags.row }}", host)
	assert.Error(t, err, "a missing map key fails rather than rendering empty")

	_, err = render(t, `{{ tag "row" }}`, host)
	assert.ErrorContains(t, err, "node has no row= tag")
}

func TestLocationTemplates(t *testing.T) {
	s := func(v string) *string { return &v }
	defaults := BmcIdentityDefaults{Rack: "{{ index .Parts 1 }}", RackUnit: "{{ index .Parts 2 }}"}

	fields, err := locationTemplates(BmcLocationBody{}, defaults)
	require.NoError(t, err)
	assert.Len(t, fields, 2, "absent fields use the config defaults")

	fields, err = locationTemplates(BmcLocationBody{Rack: s(""), Room: s("Suite 3")}, defaults)
	require.NoError(t, err)
	require.Len(t, fields, 2, "an empty field is left unchanged, even with a default")
	assert.Equal(t, "room", fields[0].name)
	assert.Equal(t, "rack_unit", fields[1].name)

	_, err = locationTemplates(BmcLocationBody{}, BmcIdentityDefaults{})
	assert.Error(t, err, "nothing to set")

	_, err = locationTemplates(BmcLocationBody{Rack: s("{{ .Name ")}, defaults)
	assert.Error(t, err, "a template that does not parse rejects the request")
}

func TestRenderLocation(t *testing.T) {
	fields, err := locationTemplates(BmcLocationBody{}, BmcIdentityDefaults{Rack: "{{ index .Parts 1 }}", RackUnit: "{{ index .Parts 2 }}"})
	require.NoError(t, err)

	u, err := renderLocation(fields, &model.Host{Name: "cpn-f01-04"}, position{})
	require.NoError(t, err)
	assert.Equal(t, "f01", *u.Rack)
	assert.True(t, u.SetRackUnit)
	assert.Equal(t, 4, *u.RackUnit)
	assert.Nil(t, u.Room)

	_, err = renderLocation(fields, &model.Host{Name: "cpn-f01-top"}, position{})
	assert.ErrorContains(t, err, "not a whole number")

	_, err = renderLocation(fields, &model.Host{Name: "switch1"}, position{})
	assert.Error(t, err)
}

func TestIdentityDataOfFailedNode(t *testing.T) {
	assert.Nil(t, assetTagData(nil))
	assert.Nil(t, locationData(nil))
}

func TestIdentityHostsOrder(t *testing.T) {
	bmc := []*model.NetInterface{{BMC: true}}
	hosts := model.HostList{
		{Name: "cpn-f01-10", Interfaces: bmc},
		{Name: "cpn-f01-9", Interfaces: bmc},
		{Name: "cpn-f01-1"},
		{Name: "cpn-f01-2", Interfaces: bmc},
	}
	var names []string
	for _, h := range identityHosts(hosts) {
		names = append(names, h.Name)
	}
	assert.Equal(t, []string{"cpn-f01-2", "cpn-f01-9", "cpn-f01-10"}, names, "natural order, skipping the host with no bmc")
}
