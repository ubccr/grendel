// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/pkg/model"
)

const sysPath = "/redfish/v1/Systems/System.Embedded.1"

func nicFixture() *fixture {
	f := dellFixture()
	f.tree[sysPath] = strings.Replace(systemDell, `"HostName"`, `"EthernetInterfaces": {"@odata.id": "`+sysPath+`/EthernetInterfaces"}, "Links": {"Chassis": [{"@odata.id": "/redfish/v1/Chassis/System.Embedded.1"}]}, "HostName"`, 1)
	f.tree[sysPath+"/EthernetInterfaces"] = `{"Members": [
		{"@odata.id": "` + sysPath + `/EthernetInterfaces/NIC.Integrated.1-1-1"},
		{"@odata.id": "` + sysPath + `/EthernetInterfaces/NIC.Slot.3-1-1"},
		{"@odata.id": "` + sysPath + `/EthernetInterfaces/NIC.Slot.3-2-1"}
	]}`
	f.tree[sysPath+"/EthernetInterfaces/NIC.Integrated.1-1-1"] = `{"@odata.id": "` + sysPath + `/EthernetInterfaces/NIC.Integrated.1-1-1", "Id": "NIC.Integrated.1-1-1", "Name": "Integrated NIC 1 Port 1", "MACAddress": "B8-CA-3A-00-00-01", "LinkStatus": "LinkDown", "SpeedMbps": 0, "InterfaceEnabled": true}`
	f.tree[sysPath+"/EthernetInterfaces/NIC.Slot.3-1-1"] = `{"@odata.id": "` + sysPath + `/EthernetInterfaces/NIC.Slot.3-1-1", "Id": "NIC.Slot.3-1-1", "Name": "Slot 3 Port 1", "MACAddress": "B8:CA:3A:00:00:10", "PermanentMACAddress": "B8:CA:3A:00:00:11", "LinkStatus": "LinkUp", "SpeedMbps": 25000, "InterfaceEnabled": true}`
	f.tree[sysPath+"/EthernetInterfaces/NIC.Slot.3-2-1"] = `{"@odata.id": "` + sysPath + `/EthernetInterfaces/NIC.Slot.3-2-1", "Id": "NIC.Slot.3-2-1", "Name": "No MAC"}`
	return f
}

func nics(t *testing.T, f *fixture) model.BmcNICList {
	t.Helper()
	srv := f.serve(t)
	r := testService(nil).NICs(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})[0]
	require.True(t, r.Success, "message was: %s", r.Message)
	require.NotNil(t, r.Data)
	return *r.Data
}

func TestNICsFromSystem(t *testing.T) {
	got := nics(t, nicFixture()).NICs

	require.Len(t, got, 2, "a port with no MAC is left out")
	assert.Equal(t, model.BmcNIC{ID: "NIC.Integrated.1-1-1", Name: "Integrated NIC 1 Port 1", MAC: "b8:ca:3a:00:00:01", Link: "LinkDown", Enabled: true}, got[0])
	assert.Equal(t, "b8:ca:3a:00:00:11", got[1].MAC, "the permanent MAC wins over a virtual one")
	assert.Equal(t, 25000, got[1].SpeedMbps)
	assert.Equal(t, "LinkUp", got[1].Link)
}

func TestNICsFallBackToNetworkAdapters(t *testing.T) {
	f := nicFixture()
	f.tree[sysPath+"/EthernetInterfaces"] = `{"Members": []}`
	f.tree["/redfish/v1/Chassis/System.Embedded.1"] = `{"@odata.id": "/redfish/v1/Chassis/System.Embedded.1", "Id": "System.Embedded.1", "NetworkAdapters": {"@odata.id": "/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters"}}`
	f.tree["/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters"] = `{"Members": [{"@odata.id": "/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Slot.3"}]}`
	f.tree["/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Slot.3"] = `{"@odata.id": "/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Slot.3", "Id": "NIC.Slot.3", "NetworkDeviceFunctions": {"@odata.id": "/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Slot.3/NetworkDeviceFunctions"}}`
	f.tree["/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Slot.3/NetworkDeviceFunctions"] = `{"Members": [{"@odata.id": "/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Slot.3/NetworkDeviceFunctions/NIC.Slot.3-1-1"}]}`
	f.tree["/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Slot.3/NetworkDeviceFunctions/NIC.Slot.3-1-1"] = `{"@odata.id": "/redfish/v1/Chassis/System.Embedded.1/NetworkAdapters/NIC.Slot.3/NetworkDeviceFunctions/NIC.Slot.3-1-1", "Id": "NIC.Slot.3-1-1", "Name": "Slot 3 Port 1", "DeviceEnabled": true, "Ethernet": {"PermanentMACAddress": "B8:CA:3A:00:00:11"}}`

	got := nics(t, f).NICs
	require.Len(t, got, 1)
	assert.Equal(t, "NIC.Slot.3-1-1", got[0].ID)
	assert.Equal(t, "b8:ca:3a:00:00:11", got[0].MAC)
	assert.True(t, got[0].Enabled, "taken from the function's DeviceEnabled")
}

func TestNICsNoneReported(t *testing.T) {
	f := nicFixture()
	f.tree[sysPath+"/EthernetInterfaces"] = `{"Members": []}`
	f.tree["/redfish/v1/Chassis/System.Embedded.1"] = `{"@odata.id": "/redfish/v1/Chassis/System.Embedded.1", "Id": "System.Embedded.1"}`
	got := nics(t, f)
	assert.NotNil(t, got.NICs, "an empty list, not null, so the UI can tell it from a node never reached")
	assert.Empty(t, got.NICs)
}

// TestNICsReadFailure keeps a bmc that could not be asked apart from a node with no NICs
func TestNICsReadFailure(t *testing.T) {
	f := nicFixture()
	f.status[sysPath+"/EthernetInterfaces"] = http.StatusInternalServerError
	f.tree["/redfish/v1/Chassis/System.Embedded.1"] = `{"@odata.id": "/redfish/v1/Chassis/System.Embedded.1", "Id": "System.Embedded.1"}`

	r := testService(nil).NICs(context.Background(), []Target{{Name: "cpn-a01-01", Addr: f.serve(t).URL}})[0]

	assert.False(t, r.Success)
	assert.Contains(t, r.Message, "reading ethernet interfaces")
	assert.Nil(t, r.Data)
}
