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

const biosSystem = `{
  "@odata.id": "/redfish/v1/Systems/System.Embedded.1",
  "Id": "System.Embedded.1",
  "Name": "System",
  "Model": "PowerEdge R650",
  "BiosVersion": "2.19.1",
  "Bios": {"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios"}
}`

const biosResource = `{
  "@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios",
  "Id": "Bios",
  "Name": "BIOS Configuration Current Settings",
  "AttributeRegistry": "BiosAttributeRegistry.v1_0_3",
  "Attributes": {"SriovGlobalEnable": "Disabled", "ProcCStates": "Enabled", "NumLock": "On", "MemTest": false, "SerialPortAddress": 3},
  "@Redfish.Settings": {
    "SettingsObject": {"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios/Settings"},
    "SupportedApplyTimes": ["Immediate", "OnReset"]
  }
}`

const biosRegistries = `{
  "@odata.id": "/redfish/v1/Registries",
  "Members": [
    {"@odata.id": "/redfish/v1/Registries/Messages"},
    {"@odata.id": "/redfish/v1/Registries/BiosAttributeRegistry"}
  ]
}`

const biosRegistryFile = `{
  "@odata.id": "/redfish/v1/Registries/BiosAttributeRegistry",
  "Id": "BiosAttributeRegistry",
  "Registry": "BiosAttributeRegistry.1.0",
  "Location": [
    {"Language": "ja", "Uri": "/registry/bios-ja.json"},
    {"Language": "en", "Uri": "/registry/bios-en.json"}
  ]
}`

const biosRegistry = `{
  "@odata.id": "/registry/bios-en.json",
  "Id": "BiosAttributeRegistry.v1_0_3",
  "Language": "en",
  "RegistryVersion": "1.0.3",
  "RegistryEntries": {
    "Attributes": [
      {"AttributeName": "SriovGlobalEnable", "DisplayName": "SR-IOV Global Enable", "MenuPath": "./IntegratedDevicesRef", "DisplayOrder": 2, "Type": "Enumeration",
       "Value": [{"ValueName": "Enabled", "ValueDisplayName": "Enabled"}, {"ValueName": "Disabled", "ValueDisplayName": "Disabled"}], "DefaultValue": "Disabled"},
      {"AttributeName": "SerialPortAddress", "DisplayName": "Serial Port Address", "MenuPath": "./SerialCommSettingsRef", "Type": "Integer", "LowerBound": 0, "UpperBound": 7, "ReadOnly": true}
    ]
  }
}`

type biosRecorder struct {
	patches []map[string]any
	headers []http.Header
	deletes []string
	staged  string
}

// biosFixture serves an iDRAC style Bios resource with a separate settings object, holding whatever the test stages. It records every PATCH and DELETE
func biosFixture(t *testing.T) (*fixture, *biosRecorder) {
	t.Helper()

	f := dellFixture()
	f.tree["/redfish/v1/"] = strings.Replace(serviceRootDell, `"JobService"`, `"Registries": {"@odata.id": "/redfish/v1/Registries"}, "JobService"`, 1)
	f.tree["/redfish/v1/Systems/System.Embedded.1"] = biosSystem
	f.tree["/redfish/v1/Systems/System.Embedded.1/Bios"] = biosResource
	f.tree["/redfish/v1/Registries"] = biosRegistries
	f.tree["/redfish/v1/Registries/Messages"] = `{"@odata.id": "/redfish/v1/Registries/Messages", "Id": "Messages", "Registry": "Base.1.0", "Location": []}`
	f.tree["/redfish/v1/Registries/BiosAttributeRegistry"] = biosRegistryFile
	f.tree["/registry/bios-en.json"] = biosRegistry

	rec := &biosRecorder{staged: `{}`}
	f.handler = func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path != "/redfish/v1/Systems/System.Embedded.1/Bios/Settings" {
			return false
		}
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("ETag", `W/"gen-1"`)
			w.Write([]byte(`{"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios/Settings", "Attributes": ` + rec.staged + `}`))
		case http.MethodPatch:
			body, _ := io.ReadAll(r.Body)
			var payload map[string]any
			_ = json.Unmarshal(body, &payload)
			rec.patches = append(rec.patches, payload)
			rec.headers = append(rec.headers, r.Header.Clone())
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			rec.deletes = append(rec.deletes, r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}
		return true
	}
	return f, rec
}

func biosTarget(t *testing.T, f *fixture) []Target {
	return []Target{{Name: "cpn-a01-01", Addr: f.serve(t).URL}}
}

func TestBiosReadsAttributesAndPending(t *testing.T) {
	f, rec := biosFixture(t)
	// iDRAC lists only the staged attributes, and one staged value equal to the applied one is not a pending change
	rec.staged = `{"SriovGlobalEnable": "Enabled", "ProcCStates": "Enabled"}`

	results := testService(nil).Bios(context.Background(), biosTarget(t, f), BiosRequest{})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)

	b := results[0].Data
	assert.Equal(t, "Disabled", b.Attributes["SriovGlobalEnable"])
	assert.Equal(t, map[string]any{"SriovGlobalEnable": "Enabled"}, b.Pending)
	assert.Equal(t, "BiosAttributeRegistry.v1_0_3", b.AttributeRegistry)
	assert.Equal(t, []string{"Immediate", "OnReset"}, b.ApplyTimes)
	assert.Equal(t, "PowerEdge R650", b.Model)
	assert.Equal(t, "2.19.1", b.BIOSVersion)
}

func TestBiosFiltersNames(t *testing.T) {
	f, rec := biosFixture(t)
	rec.staged = `{"SriovGlobalEnable": "Enabled", "NumLock": "Off"}`

	results := testService(nil).Bios(context.Background(), biosTarget(t, f), BiosRequest{Names: []string{"NumLock", "Missing"}})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, map[string]any{"NumLock": "On"}, results[0].Data.Attributes)
	assert.Equal(t, map[string]any{"NumLock": "Off"}, results[0].Data.Pending)
}

func TestSetBiosSendsChangesWithApplyTime(t *testing.T) {
	f, rec := biosFixture(t)

	results := testService(nil).SetBios(context.Background(), biosTarget(t, f), SetBiosRequest{
		// strings, as the cli sends them, for an enumeration, a number and a boolean, plus one value that already matches
		Attributes: map[string]any{"SriovGlobalEnable": "Enabled", "SerialPortAddress": "4", "MemTest": "true", "NumLock": "On"},
		ApplyTime:  "OnReset",
	})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Contains(t, results[0].Message, "staged 4 bios attribute(s), apply time OnReset")

	require.Len(t, rec.patches, 1)
	assert.Equal(t, map[string]any{"SriovGlobalEnable": "Enabled", "SerialPortAddress": float64(4), "MemTest": true}, rec.patches[0]["Attributes"],
		"only changed values are sent, coerced to the type the bmc reports")
	assert.Equal(t, map[string]any{"ApplyTime": "OnReset"}, rec.patches[0]["@Redfish.SettingsApplyTime"])
	assert.Equal(t, `W/"gen-1"`, rec.headers[0].Get("If-Match"))
}

func TestSetBiosRefusesBeforeSending(t *testing.T) {
	cases := map[string]struct {
		req  SetBiosRequest
		want string
	}{
		"unknown attribute":     {SetBiosRequest{Attributes: map[string]any{"NoSuchThing": "x"}}, "unknown bios attribute NoSuchThing"},
		"not a number":          {SetBiosRequest{Attributes: map[string]any{"SerialPortAddress": "four"}}, `"four" is not a number`},
		"not a boolean":         {SetBiosRequest{Attributes: map[string]any{"MemTest": "maybe"}}, `"maybe" is not true or false`},
		"string given a number": {SetBiosRequest{Attributes: map[string]any{"NumLock": float64(1)}}, "expects a string"},
		"apply time not offered": {SetBiosRequest{Attributes: map[string]any{"NumLock": "Off"}, ApplyTime: "AtMaintenanceWindowStart"},
			"does not accept apply time AtMaintenanceWindowStart, it accepts Immediate, OnReset"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f, rec := biosFixture(t)

			results := testService(nil).SetBios(context.Background(), biosTarget(t, f), tc.req)
			assert.False(t, results[0].Success)
			assert.Contains(t, results[0].Message, tc.want)
			assert.Empty(t, rec.patches, "nothing is sent when any value is refused")
		})
	}
}

func TestSetBiosAlreadySet(t *testing.T) {
	f, rec := biosFixture(t)
	rec.staged = `{"SriovGlobalEnable": "Enabled"}`

	// one matches what is applied and the other what is already staged
	results := testService(nil).SetBios(context.Background(), biosTarget(t, f), SetBiosRequest{
		Attributes: map[string]any{"NumLock": "On", "SriovGlobalEnable": "Enabled"},
	})
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "already set, nothing sent", results[0].Message)
	assert.Empty(t, rec.patches)
}

func TestDiscardPendingBios(t *testing.T) {
	f, rec := biosFixture(t)
	rec.staged = `{"SriovGlobalEnable": "Enabled"}`

	results := testService(nil).DiscardPendingBios(context.Background(), biosTarget(t, f))
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "discarded 1 staged bios change(s)", results[0].Message)
	assert.Equal(t, []string{"/redfish/v1/Systems/System.Embedded.1/Bios/Settings"}, rec.deletes)
}

func TestDiscardPendingBiosNothingStaged(t *testing.T) {
	f, rec := biosFixture(t)

	results := testService(nil).DiscardPendingBios(context.Background(), biosTarget(t, f))
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "no staged bios changes", results[0].Message)
	assert.Empty(t, rec.deletes)
}

func TestBiosRegistry(t *testing.T) {
	f, _ := biosFixture(t)
	targets := biosTarget(t, f)
	svc := testService(nil)

	results := svc.BiosRegistry(context.Background(), targets)
	require.True(t, results[0].Success, "message was: %s", results[0].Message)

	reg := results[0].Data
	assert.True(t, reg.Available)
	assert.Equal(t, "en", reg.Language, "the english copy is preferred")
	require.Len(t, reg.Attributes, 2)

	sriov := reg.Attributes[0]
	assert.Equal(t, "SriovGlobalEnable", sriov.Name)
	assert.Equal(t, "SR-IOV Global Enable", sriov.DisplayName)
	assert.Equal(t, "Enumeration", sriov.Type)
	assert.Len(t, sriov.Values, 2)

	port := reg.Attributes[1]
	assert.True(t, port.ReadOnly)
	require.NotNil(t, port.UpperBound)
	assert.EqualValues(t, 7, *port.UpperBound)

	// a second node of the same model and bios version is served from the cache
	svc.BiosRegistry(context.Background(), targets)
	assert.Equal(t, 1, f.hits["/registry/bios-en.json"])
}

// TestBiosRegistryIdracNaming is the layout an iDRAC serves, and a mockup made from one: the Bios resource, the registry file id and the file's Registry each carry a different version
func TestBiosRegistryIdracNaming(t *testing.T) {
	f, _ := biosFixture(t)
	f.tree["/redfish/v1/Registries"] = strings.Replace(biosRegistries, "/redfish/v1/Registries/BiosAttributeRegistry", "/redfish/v1/Registries/BiosAttributeRegistry.v1_0_0", 1)
	f.tree["/redfish/v1/Registries/BiosAttributeRegistry.v1_0_0"] = `{
  "@odata.id": "/redfish/v1/Registries/BiosAttributeRegistry.v1_0_0",
  "Id": "BiosAttributeRegistry.v1_0_0",
  "Registry": "BiosAttributeRegistry1.0",
  "Location": [{"Language": "en", "Uri": "/registry/bios-en.json"}]
}`

	results := testService(nil).BiosRegistry(context.Background(), biosTarget(t, f))
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.True(t, results[0].Data.Available)
	assert.Len(t, results[0].Data.Attributes, 2)
}

func TestRegistryBase(t *testing.T) {
	for _, name := range []string{"BiosAttributeRegistry.v1_0_3", "BiosAttributeRegistry.v1_0_0", "BiosAttributeRegistry1.0", "BiosAttributeRegistry.1.0", "BiosAttributeRegistry"} {
		assert.Equal(t, "BiosAttributeRegistry", registryBase(name), name)
	}
	assert.NotEqual(t, registryBase("BiosAttributeRegistry.v1_0_3"), registryBase("ManagerAttributeRegistry.v1_0_0"))
}

func TestBiosRegistryOnlyPublished(t *testing.T) {
	f, _ := biosFixture(t)
	f.tree["/redfish/v1/Registries/BiosAttributeRegistry"] = strings.Replace(biosRegistryFile,
		`[
    {"Language": "ja", "Uri": "/registry/bios-ja.json"},
    {"Language": "en", "Uri": "/registry/bios-en.json"}
  ]`, `[{"Language": "en", "PublicationUri": "https://example.com/bios.json"}]`, 1)

	results := testService(nil).BiosRegistry(context.Background(), biosTarget(t, f))
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.False(t, results[0].Data.Available)
	assert.Empty(t, results[0].Data.Attributes)
}

func TestRegistryCacheEvictsOldest(t *testing.T) {
	c := newRegistryCache(2)
	c.put("a", nil)
	c.put("b", nil)
	c.get("a")
	c.put("c", nil)

	_, a := c.get("a")
	_, b := c.get("b")
	_, cc := c.get("c")
	assert.True(t, a, "a was used most recently before c arrived")
	assert.False(t, b)
	assert.True(t, cc)
}
