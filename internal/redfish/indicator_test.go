// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const indicatorChassis = `{
  "@odata.id": "/redfish/v1/Chassis/System.Embedded.1",
  "Id": "System.Embedded.1",
  "Name": "Chassis",
  "IndicatorLED": "Off"
}`

// indicatorFixture serves a system with the given indicator properties, linked to a chassis that reports IndicatorLED, and records every PATCH by path
func indicatorFixture(systemProps string) (*fixture, map[string]map[string]any) {
	f := dellFixture()
	system := `{
  "@odata.id": "/redfish/v1/Systems/System.Embedded.1",
  "Id": "System.Embedded.1",
  "Name": "System",
  "Links": {"Chassis": [{"@odata.id": "/redfish/v1/Chassis/System.Embedded.1"}]}` + systemProps + `
}`
	f.tree["/redfish/v1/Systems/System.Embedded.1"] = system
	f.tree["/redfish/v1/Chassis/System.Embedded.1"] = indicatorChassis

	patches := map[string]map[string]any{}
	f.handler = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method != http.MethodPatch {
			return false
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		patches[r.URL.Path] = payload
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return f, patches
}

func TestIndicator(t *testing.T) {
	cases := []struct {
		name        string
		systemProps string
		on          bool
		path        string
		want        map[string]any
	}{
		{"location indicator active on the system", `, "LocationIndicatorActive": false, "IndicatorLED": "Off"`, true,
			"/redfish/v1/Systems/System.Embedded.1", map[string]any{"LocationIndicatorActive": true}},
		{"turned off", `, "LocationIndicatorActive": true`, false,
			"/redfish/v1/Systems/System.Embedded.1", map[string]any{"LocationIndicatorActive": false}},
		{"only the deprecated led", `, "IndicatorLED": "Off"`, true,
			"/redfish/v1/Systems/System.Embedded.1", map[string]any{"IndicatorLED": "Blinking"}},
		{"neither on the system falls back to the chassis", ``, true,
			"/redfish/v1/Chassis/System.Embedded.1", map[string]any{"IndicatorLED": "Blinking"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, patches := indicatorFixture(tc.systemProps)
			srv := f.serve(t)

			results := testService(nil).Indicator(context.Background(),
				[]Target{{Name: "cpn-a01-01", Addr: srv.URL}}, IndicatorRequest{On: tc.on})

			require.Len(t, results, 1)
			require.True(t, results[0].Success, "message was: %s", results[0].Message)
			require.Len(t, patches, 1, "exactly one resource is written")
			assert.Equal(t, tc.want, patches[tc.path])
		})
	}
}

func TestIndicatorNotReported(t *testing.T) {
	f, patches := indicatorFixture(``)
	f.tree["/redfish/v1/Chassis/System.Embedded.1"] = strings.Replace(indicatorChassis, `,
  "IndicatorLED": "Off"`, "", 1)
	srv := f.serve(t)

	results := testService(nil).Indicator(context.Background(),
		[]Target{{Name: "cpn-a01-01", Addr: srv.URL}}, IndicatorRequest{On: true})

	require.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.Contains(t, results[0].Message, "no location indicator")
	assert.Empty(t, patches)
}

func TestStatusIndicatorState(t *testing.T) {
	cases := map[string]string{
		`, "LocationIndicatorActive": true`:  "On",
		`, "LocationIndicatorActive": false`: "Off",
		`, "IndicatorLED": "Blinking"`:       "On",
		`, "IndicatorLED": "Off"`:            "Off",
		``:                                   "",
	}
	for props, want := range cases {
		f, _ := indicatorFixture(props)
		srv := f.serve(t)

		results := testService(nil).Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})
		require.True(t, results[0].Success, "message was: %s", results[0].Message)
		assert.Equal(t, want, results[0].Data.LocationIndicator, "system props %q", props)
	}
}
