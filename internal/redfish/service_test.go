// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixture is a Redfish tree keyed by request path. gofish only needs an endpoint that speaks http, so an httptest server is the seam that lets the whole stack be tested, gofish's own parsing included, without a mock in front of it
type fixture struct {
	tree    map[string]string
	status  map[string]int
	delay   map[string]time.Duration
	hits    map[string]int
	handler func(w http.ResponseWriter, r *http.Request) bool
	// collections are writable collections by uri, holding each member's body by its uri. A POST adds a member and a DELETE removes one, so the next GET sees the change
	collections map[string]map[string]string
	// writes records every request that is not a GET, in order
	writes []fixtureWrite
	nextID int
}

type fixtureWrite struct {
	Method string
	Path   string
	Body   map[string]any
}

func newFixture() *fixture {
	return &fixture{
		tree:        map[string]string{},
		status:      map[string]int{},
		delay:       map[string]time.Duration{},
		hits:        map[string]int{},
		collections: map[string]map[string]string{},
	}
}

// member puts a member in a collection, as if it was there before the test, creating the collection if need be
func (f *fixture) member(coll, id string, body map[string]any) {
	if f.collections[coll] == nil {
		f.collections[coll] = map[string]string{}
	}
	uri := coll + "/" + id
	full := map[string]any{"@odata.id": uri, "Id": id}
	maps.Copy(full, body)
	raw, _ := json.Marshal(full)
	f.collections[coll][uri] = string(raw)
}

// written returns the bodies sent to a path with a method, in order
func (f *fixture) written(method, path string) []map[string]any {
	var out []map[string]any
	for _, w := range f.writes {
		if w.Method == method && w.Path == path {
			out = append(out, w.Body)
		}
	}
	return out
}

// serveCollection answers for the writable collections and their members, and reports whether it did
func (f *fixture) serveCollection(w http.ResponseWriter, r *http.Request, body map[string]any) bool {
	p := r.URL.Path
	if coll, ok := f.collections[p]; ok {
		switch r.Method {
		case http.MethodGet:
			refs := []string{}
			for _, uri := range slices.Sorted(maps.Keys(coll)) {
				refs = append(refs, `{"@odata.id": "`+uri+`"}`)
			}
			w.Write([]byte(`{"@odata.id": "` + p + `", "Members": [` + strings.Join(refs, ",") + `]}`))
		case http.MethodPost:
			f.nextID++
			id := strconv.Itoa(f.nextID)
			f.member(p, id, body)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(coll[p+"/"+id]))
		}
		return true
	}
	for _, coll := range f.collections {
		member, ok := coll[p]
		if !ok {
			continue
		}
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte(member))
		case http.MethodDelete:
			delete(coll, p)
			w.WriteHeader(http.StatusNoContent)
		}
		return true
	}
	return false
}

func (f *fixture) serve(t *testing.T) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.hits[r.URL.Path]++

		var sent map[string]any
		if r.Method != http.MethodGet {
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &sent)
			r.Body = io.NopCloser(bytes.NewReader(raw))
			f.writes = append(f.writes, fixtureWrite{Method: r.Method, Path: r.URL.Path, Body: sent})
		}

		if f.handler != nil && f.handler(w, r) {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if f.serveCollection(w, r, sent) {
			return
		}

		if d, ok := f.delay[r.URL.Path]; ok {
			select {
			case <-time.After(d):
			case <-r.Context().Done():
				return
			}
		}

		if code, ok := f.status[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			if body, ok := f.tree[r.URL.Path]; ok {
				w.Write([]byte(body))
			}
			return
		}

		body, ok := f.tree[r.URL.Path]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":{"code":"Base.1.0.ResourceMissing","message":"resource not found"}}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	return srv
}

func testService(cfg func(*Config)) *Service {
	c := Config{
		Credentials: []Credential{{User: "grendel", Password: "secret"}},
		BasicAuth:   true,
		Insecure:    true,
		Fanout:      4,
		NodeTimeout: 5 * time.Second,
	}
	if cfg != nil {
		cfg(&c)
	}
	return New(c)
}

const serviceRootDell = `{
  "@odata.id": "/redfish/v1/",
  "@odata.type": "#ServiceRoot.v1_5_0.ServiceRoot",
  "Id": "RootService",
  "Name": "Root Service",
  "Vendor": "Dell",
  "Systems": {"@odata.id": "/redfish/v1/Systems"},
  "Managers": {"@odata.id": "/redfish/v1/Managers"},
  "JobService": {"@odata.id": "/redfish/v1/JobService"}
}`

const serviceRootGeneric = `{
  "@odata.id": "/redfish/v1/",
  "@odata.type": "#ServiceRoot.v1_5_0.ServiceRoot",
  "Id": "RootService",
  "Name": "Root Service",
  "Vendor": "Contoso",
  "Systems": {"@odata.id": "/redfish/v1/Systems"},
  "Managers": {"@odata.id": "/redfish/v1/Managers"}
}`

const systemsCollection = `{
  "@odata.id": "/redfish/v1/Systems",
  "Name": "Computer System Collection",
  "Members@odata.count": 1,
  "Members": [{"@odata.id": "/redfish/v1/Systems/System.Embedded.1"}]
}`

const emptyCollection = `{
  "@odata.id": "/redfish/v1/Systems",
  "Name": "Computer System Collection",
  "Members@odata.count": 0,
  "Members": []
}`

const systemDell = `{
  "@odata.id": "/redfish/v1/Systems/System.Embedded.1",
  "@odata.type": "#ComputerSystem.v1_12_0.ComputerSystem",
  "Id": "System.Embedded.1",
  "Name": "System",
  "HostName": "cpn-a01-01",
  "BiosVersion": "2.19.1",
  "SKU": "ABCD123",
  "Manufacturer": "Dell Inc.",
  "Model": "PowerEdge R650",
  "PowerState": "On",
  "Status": {"Health": "OK", "State": "Enabled"},
  "MemorySummary": {"TotalSystemMemoryGiB": 512},
  "ProcessorSummary": {"LogicalProcessorCount": 96, "Count": 2, "CoreCount": 48, "Model": "Intel(R) Xeon(R) Gold 6342 CPU @ 2.80GHz"},
  "BootProgress": {"LastState": "OSRunning", "LastStateTime": "2026-11-09T08:14:02-05:00"},
  "Boot": {
    "BootNext": "",
    "BootOrder": ["NIC.Slot.1-1", "HardDisk.List.1-1"],
    "BootSourceOverrideTarget": "Pxe",
    "BootSourceOverrideEnabled": "Once",
    "BootSourceOverrideMode": "UEFI"
  },
  "Oem": {"Dell": {"DellSystem": {"SystemID": 2321, "NodeID": "ABCD123"}}}
}`

func dellFixture() *fixture {
	f := newFixture()
	f.tree["/redfish/v1/"] = serviceRootDell
	f.tree["/redfish/v1/Systems"] = systemsCollection
	f.tree["/redfish/v1/Systems/System.Embedded.1"] = systemDell
	return f
}

func TestStatusParsesSystem(t *testing.T) {
	srv := dellFixture().serve(t)

	results := testService(nil).Status(context.Background(),
		[]Target{{Name: "cpn-a01-01", Addr: srv.URL}})

	require.Len(t, results, 1)
	r := results[0]
	require.True(t, r.Success, "message was: %s", r.Message)
	require.NotNil(t, r.Data)

	assert.Equal(t, "cpn-a01-01", r.Data.Name)
	assert.Equal(t, "2.19.1", r.Data.BIOSVersion)
	assert.Equal(t, "ABCD123", r.Data.SerialNumber)
	assert.Equal(t, "PowerEdge R650", r.Data.Model)
	assert.Equal(t, "On", r.Data.PowerStatus)
	assert.Equal(t, "OK", r.Data.Health)
	assert.EqualValues(t, 512, r.Data.TotalMemory)
	assert.Equal(t, 96, r.Data.ProcessorCount)
	assert.Equal(t, []string{"NIC.Slot.1-1", "HardDisk.List.1-1"}, r.Data.BootOrder)
	assert.Equal(t, "Intel(R) Xeon(R) Gold 6342 CPU @ 2.80GHz", r.Data.ProcessorModel)
	assert.Equal(t, 2, r.Data.ProcessorSockets)
	assert.Equal(t, 48, r.Data.ProcessorCores)
	assert.Equal(t, "OSRunning", r.Data.BootProgress)
	assert.Equal(t, "2026-11-09T08:14:02-05:00", r.Data.BootProgressTime)
	assert.Equal(t, "Pxe", r.Data.BootOverrideTarget)
	assert.Equal(t, "Once", r.Data.BootOverrideEnabled)
	assert.Equal(t, "UEFI", r.Data.BootMode)

}

// TestStatusOverPlainHTTP is the whole bmc.scheme path: a bare address with no scheme, as TargetsFromHosts produces from a bmc interface, reaching a server that speaks plain http. This is what pointing the daemon at a redfish mockup server looks like
// TestStatusVendorBootProgress is a vendor boot state, which Redfish reports as LastState OEM with the real name beside it
func TestStatusVendorBootProgress(t *testing.T) {
	f := dellFixture()
	f.tree["/redfish/v1/Systems/System.Embedded.1"] = strings.Replace(systemDell,
		`"LastState": "OSRunning"`, `"LastState": "OEM", "OEMLastState": "UEFIShellRunning"`, 1)
	srv := f.serve(t)

	r := testService(nil).Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})[0]

	require.True(t, r.Success, "message was: %s", r.Message)
	assert.Equal(t, "UEFIShellRunning", r.Data.BootProgress)
}

func TestStatusOverPlainHTTP(t *testing.T) {
	srv := dellFixture().serve(t)
	addr := strings.TrimPrefix(srv.URL, "http://")
	require.NotContains(t, addr, "://", "the target must carry no scheme, or it would pass through untouched")

	svc := testService(func(c *Config) { c.Scheme = schemeHTTP })

	results := svc.Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: addr}})

	require.Len(t, results, 1)
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "PowerEdge R650", results[0].Data.Model)
}

func TestStatusEmptySystemCollection(t *testing.T) {
	f := dellFixture()
	f.tree["/redfish/v1/Systems"] = emptyCollection

	srv := f.serve(t)

	results := testService(nil).Status(context.Background(),
		[]Target{{Name: "cpn-a01-01", Addr: srv.URL}})

	require.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.Contains(t, results[0].Message, "no computer systems")
	assert.Nil(t, results[0].Data)
}

func TestStatusUnauthorized(t *testing.T) {
	f := dellFixture()
	f.status["/redfish/v1/"] = http.StatusUnauthorized
	f.tree["/redfish/v1/"] = dellAuthError

	srv := f.serve(t)

	svc := testService(func(c *Config) {
		c.Credentials = []Credential{
			{User: "grendel", Password: "wrong"},
			{User: "root", Password: "calvin"},
		}
	})

	results := svc.Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})

	require.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.Contains(t, results[0].Message, "tried grendel, root", "the operator has to see which logins were attempted")
	assert.Contains(t, results[0].Message, "invalid username")
	assert.Equal(t, 2, f.hits["/redfish/v1/"], "every configured credential should be tried before giving up")
}

func TestStatusUnreachable(t *testing.T) {
	srv := dellFixture().serve(t)
	addr := srv.URL
	srv.Close()

	results := testService(nil).Status(context.Background(),
		[]Target{{Name: "cpn-a01-01", Addr: addr}})

	require.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.Contains(t, results[0].Message, "connection refused")
}

func TestStatusNodeTimeout(t *testing.T) {
	f := dellFixture()
	f.delay["/redfish/v1/"] = 2 * time.Second

	srv := f.serve(t)

	svc := testService(func(c *Config) { c.NodeTimeout = 100 * time.Millisecond })

	begin := time.Now()
	results := svc.Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})

	require.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.Less(t, time.Since(begin), time.Second, "the node deadline must actually cut the request short")
}

// TestJobsNoJobService covers a BMC that works but does not do jobs, where gofish hands back a nil job service
func TestJobsNoJobService(t *testing.T) {
	f := newFixture()
	f.tree["/redfish/v1/"] = serviceRootGeneric
	f.tree["/redfish/v1/Systems"] = systemsCollection

	srv := f.serve(t)

	results := testService(nil).Jobs(context.Background(),
		[]Target{{Name: "sw-a01-01", Addr: srv.URL}})

	require.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.Contains(t, results[0].Message, "does not implement a redfish job service")
}

// TestMixedNodesetReportsEveryNode is the end to end version of the invariant that every target gets a row
func TestMixedNodesetReportsEveryNode(t *testing.T) {
	good := dellFixture().serve(t)

	dead := dellFixture().serve(t)
	deadAddr := dead.URL
	dead.Close()

	empty := dellFixture()
	empty.tree["/redfish/v1/Systems"] = emptyCollection
	emptySrv := empty.serve(t)

	targets := []Target{
		{Name: "cpn-good", Addr: good.URL},
		{Name: "cpn-dead", Addr: deadAddr},
		{Name: "cpn-empty", Addr: emptySrv.URL},
	}

	results := testService(nil).Status(context.Background(), targets)

	require.Len(t, results, 3)
	assert.Equal(t, "cpn-good", results[0].Node)
	assert.Equal(t, "cpn-dead", results[1].Node)
	assert.Equal(t, "cpn-empty", results[2].Node)

	assert.True(t, results[0].Success)
	assert.False(t, results[1].Success)
	assert.False(t, results[2].Success)
}
