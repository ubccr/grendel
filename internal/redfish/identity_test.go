// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/pkg/model"
)

const identityChassis = `{
  "@odata.id": "/redfish/v1/Chassis/System.Embedded.1",
  "Id": "System.Embedded.1",
  "Name": "Chassis",
  "Location": {
    "Info": "VTC;Suite 3;;f01;4",
    "InfoFormat": "DataCenter;RoomName;Aisle;RackName;RackSlot",
    "Placement": {"Rack": "f01", "RackOffset": 4, "RackOffsetUnits": "EIA_310", "Row": ""},
    "PostalAddress": {"Building": "VTC", "Room": "Suite 3"}
  }
}`

const identityManagers = `{"Members": [{"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1"}]}`

const identityManager = `{"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1", "Id": "iDRAC.Embedded.1", "Name": "Manager"}`

const identityAttributes = `{"Attributes": {
  "ServerTopology.1.DataCenterName": "VTC",
  "ServerTopology.1.RoomName": "Suite 3",
  "ServerTopology.1.AisleName": "",
  "ServerTopology.1.RackName": "f01",
  "ServerTopology.1.RackSlot": 4
}}`

// identityFixture is a Dell system with an asset tag and a chassis location, and with the iDRAC topology attributes when attributes is true
func identityFixture(attributes bool) (*fixture, map[string]map[string]any) {
	f, patches := indicatorFixture(`, "AssetTag": "OLD-1", "SKU": "ABCD123"`)
	f.tree["/redfish/v1/Chassis/System.Embedded.1"] = identityChassis
	f.tree["/redfish/v1/Managers"] = identityManagers
	f.tree["/redfish/v1/Managers/iDRAC.Embedded.1"] = identityManager
	if attributes {
		f.tree["/redfish/v1/Managers/iDRAC.Embedded.1/Attributes"] = identityAttributes
	}
	return f, patches
}

func ptr[T any](v T) *T { return &v }

func TestSetAssetTag(t *testing.T) {
	cases := []struct {
		name    string
		req     AssetTagRequest
		ok      bool
		message string
		patched any
	}{
		{"by node", AssetTagRequest{ByNode: map[string]string{"cpn-a01-01": "NEW-1"}}, true, "asset tag set", "NEW-1"},
		{"by serial", AssetTagRequest{BySerial: map[string]string{"ABCD123": "NEW-2"}}, true, "asset tag set", "NEW-2"},
		{"serial missing from the csv", AssetTagRequest{BySerial: map[string]string{"ZZZ": "NEW-2"}}, false, "no row for serial ABCD123", nil},
		{"already set", AssetTagRequest{ByNode: map[string]string{"cpn-a01-01": "OLD-1"}}, true, "already set", nil},
		{"dry run", AssetTagRequest{ByNode: map[string]string{"cpn-a01-01": "NEW-1"}, DryRun: true}, true, "would change, dry run", nil},
		{"invalid", AssetTagRequest{Invalid: map[string]error{"cpn-a01-01": errors.New("template: bad")}}, false, "template: bad", nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, patches := identityFixture(false)
			srv := f.serve(t)

			r := testService(nil).SetAssetTag(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}}, tc.req)[0]

			require.Equal(t, tc.ok, r.Success, "message was: %s", r.Message)
			assert.Contains(t, r.Message, tc.message)
			got := patches["/redfish/v1/Systems/System.Embedded.1"]
			if tc.patched == nil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, map[string]any{"AssetTag": tc.patched}, got)
				assert.Equal(t, "OLD-1", r.Data.Current)
				assert.Equal(t, "ABCD123", r.Data.Serial)
			}
		})
	}
}

func TestSetLocationDellAttributes(t *testing.T) {
	f, patches := identityFixture(true)
	srv := f.serve(t)

	req := LocationRequest{ByNode: map[string]LocationUpdate{"cpn-a01-01": {Rack: ptr("f02"), RackUnit: ptr(7), SetRackUnit: true}}}
	r := testService(nil).SetLocation(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}}, req)[0]

	require.True(t, r.Success, "message was: %s", r.Message)
	assert.Equal(t, map[string]any{"Attributes": map[string]any{
		"ServerTopology.1.RackName": "f02",
		"ServerTopology.1.RackSlot": float64(7),
	}}, patches["/redfish/v1/Managers/iDRAC.Embedded.1/Attributes"])
	assert.Equal(t, "f01", r.Data.Current.Rack)
	assert.Equal(t, "f02", r.Data.Desired.Rack)
	assert.Equal(t, "VTC", r.Data.Desired.Datacenter, "fields not given are kept")
	assert.True(t, r.Data.Changed)
}

func TestSetLocationChassis(t *testing.T) {
	f, patches := identityFixture(false)
	srv := f.serve(t)

	req := LocationRequest{ByNode: map[string]LocationUpdate{"cpn-a01-01": {Room: ptr("Suite 4"), Rack: ptr("f02"), RackUnit: ptr(7), SetRackUnit: true}}}
	r := testService(nil).SetLocation(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}}, req)[0]

	require.True(t, r.Success, "message was: %s", r.Message)
	assert.Equal(t, map[string]any{"Location": map[string]any{
		"Placement":     map[string]any{"Rack": "f02", "RackOffset": float64(7), "RackOffsetUnits": "EIA_310"},
		"PostalAddress": map[string]any{"Room": "Suite 4"},
	}}, patches["/redfish/v1/Chassis/System.Embedded.1"])
}

func TestSetLocationUnchangedAndDryRun(t *testing.T) {
	f, patches := identityFixture(true)
	srv := f.serve(t)
	targets := []Target{{Name: "cpn-a01-01", Addr: srv.URL}}

	same := LocationRequest{ByNode: map[string]LocationUpdate{"cpn-a01-01": {Rack: ptr("f01"), RackUnit: ptr(4), SetRackUnit: true}}}
	r := testService(nil).SetLocation(context.Background(), targets, same)[0]
	require.True(t, r.Success, "message was: %s", r.Message)
	assert.Equal(t, "already set", r.Message)
	assert.False(t, r.Data.Changed)

	dry := LocationRequest{ByNode: map[string]LocationUpdate{"cpn-a01-01": {Rack: ptr("f09")}}, DryRun: true}
	r = testService(nil).SetLocation(context.Background(), targets, dry)[0]
	require.True(t, r.Success, "message was: %s", r.Message)
	assert.True(t, r.Data.Changed)
	assert.Empty(t, patches)
}

func TestInfoLocation(t *testing.T) {
	got := infoLocation("VTC;Suite 3;A;f01;04", "DataCenter;RoomName;Aisle;RackName;RackSlot")
	assert.Equal(t, model.BmcLocation{Datacenter: "VTC", Room: "Suite 3", Row: "A", Rack: "f01", RackUnit: ptr(4)}, got)
}

func TestStatusReadsIdentity(t *testing.T) {
	f, _ := identityFixture(false)
	srv := f.serve(t)

	r := testService(nil).Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})[0]
	require.True(t, r.Success, "message was: %s", r.Message)
	assert.Equal(t, "OLD-1", r.Data.AssetTag)
	require.NotNil(t, r.Data.Location)
	assert.Equal(t, "f01", r.Data.Location.Rack)
	assert.Equal(t, 4, *r.Data.Location.RackUnit)
}
