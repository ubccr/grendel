// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func systemWith(fields string) string {
	return fmt.Sprintf(`{
	  "@odata.id": "/redfish/v1/Systems/1",
	  "@odata.type": "#ComputerSystem.v1_12_0.ComputerSystem",
	  "Id": "1",
	  "Name": "System",
	  %s
	}`, fields)
}

func vendorFixture(root, system string) *fixture {
	f := newFixture()
	f.tree["/redfish/v1/"] = root
	f.tree["/redfish/v1/Systems"] = `{
	  "@odata.id": "/redfish/v1/Systems",
	  "Members@odata.count": 1,
	  "Members": [{"@odata.id": "/redfish/v1/Systems/1"}]
	}`
	f.tree["/redfish/v1/Systems/1"] = system
	return f
}

func TestSerialNumberByVendor(t *testing.T) {
	tests := []struct {
		name   string
		root   string
		fields string
		want   string
	}{
		{"dell uses the service tag, not the board serial", serviceRootDell, `"SKU": "5BWFB03", "SerialNumber": "CNWS3009A7009B"`, "5BWFB03"},
		{"dell with no sku falls back", serviceRootDell, `"SerialNumber": "CNWS3009A7009B"`, "CNWS3009A7009B"},
		{"others use the serial, a supermicro sku is not one", serviceRootGeneric, `"SKU": "0x1D1415D9", "SerialNumber": "S514359X4916804"`, "S514359X4916804"},
		{"a placeholder sku is never reported", serviceRootGeneric, `"SKU": "Default string"`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := vendorFixture(tt.root, systemWith(tt.fields)).serve(t)

			results := testService(nil).Status(context.Background(), []Target{{Name: "n", Addr: srv.URL}})

			require.True(t, results[0].Success, "message was: %s", results[0].Message)
			assert.Equal(t, tt.want, results[0].Data.SerialNumber)
		})
	}
}

func TestOEMBlockOnlyOnDell(t *testing.T) {
	dell := dellFixture().serve(t)
	other := vendorFixture(serviceRootGeneric, systemWith(`"SerialNumber": "S1"`)).serve(t)

	svc := testService(nil)

	dellResults := svc.Status(context.Background(), []Target{{Name: "n", Addr: dell.URL}})
	require.True(t, dellResults[0].Success)
	assert.NotNil(t, dellResults[0].Data.OEMDell)

	otherResults := svc.Status(context.Background(), []Target{{Name: "n", Addr: other.URL}})
	require.True(t, otherResults[0].Success)
	assert.Nil(t, otherResults[0].Data.OEMDell, "a non Dell row must serialize null rather than an empty block")
}

// logTree adds a log service collection to a resource path, each given as id and LogEntryType, with each log holding one entry named after it
func logTree(f *fixture, parent string, logs ...[2]string) {
	members := make([]string, 0, len(logs))
	for _, l := range logs {
		id, logType := l[0], l[1]
		uri := parent + "/LogServices/" + id
		members = append(members, fmt.Sprintf(`{"@odata.id": %q}`, uri))

		typeField := ""
		if logType != "" {
			typeField = fmt.Sprintf(`"LogEntryType": %q,`, logType)
		}

		f.tree[uri] = fmt.Sprintf(`{
		  "@odata.id": %q,
		  "@odata.type": "#LogService.v1_1_0.LogService",
		  "Id": %q,
		  "Name": %q,
		  %s
		  "Entries": {"@odata.id": %q},
		  "Actions": {"#LogService.ClearLog": {"target": %q}}
		}`, uri, id, id, typeField, uri+"/Entries", uri+"/Actions/LogService.ClearLog")

		entry := uri + "/Entries/1"
		f.tree[uri+"/Entries"] = fmt.Sprintf(`{
		  "@odata.id": %q,
		  "Members@odata.count": 1,
		  "Members": [{"@odata.id": %q}]
		}`, uri+"/Entries", entry)
		f.tree[entry] = fmt.Sprintf(`{
		  "@odata.id": %q,
		  "@odata.type": "#LogEntry.v1_4_0.LogEntry",
		  "Id": %q,
		  "Name": "Entry",
		  "Created": "2026-02-23T15:44:58-06:00",
		  "EntryType": "SEL",
		  "Message": "entry"
		}`, entry, id)
	}

	f.tree[parent+"/LogServices"] = fmt.Sprintf(`{
	  "@odata.id": %q,
	  "Members@odata.count": %d,
	  "Members": [%s]
	}`, parent+"/LogServices", len(logs), strings.Join(members, ","))
}

const (
	managerPath = "/redfish/v1/Managers/1"
	systemPath  = "/redfish/v1/Systems/1"
)

// logVendorFixture is a service with one manager and one system, each linking their log services when they have any
func logVendorFixture(managerLogs, systemLogs [][2]string) *fixture {
	systemFields := `"SerialNumber": "S1"`
	if systemLogs != nil {
		systemFields += fmt.Sprintf(`, "LogServices": {"@odata.id": %q}`, systemPath+"/LogServices")
	}
	f := vendorFixture(serviceRootGeneric, systemWith(systemFields))

	f.tree["/redfish/v1/Managers"] = fmt.Sprintf(`{
	  "@odata.id": "/redfish/v1/Managers",
	  "Members@odata.count": 1,
	  "Members": [{"@odata.id": %q}]
	}`, managerPath)

	managerLink := ""
	if managerLogs != nil {
		managerLink = fmt.Sprintf(`, "LogServices": {"@odata.id": %q}`, managerPath+"/LogServices")
		logTree(f, managerPath, managerLogs...)
	}
	f.tree[managerPath] = fmt.Sprintf(`{
	  "@odata.id": %q,
	  "@odata.type": "#Manager.v1_5_0.Manager",
	  "Id": "1",
	  "Name": "Manager"%s
	}`, managerPath, managerLink)

	if systemLogs != nil {
		logTree(f, systemPath, systemLogs...)
	}

	return f
}

func TestSELLocationByVendor(t *testing.T) {
	tests := []struct {
		name    string
		manager [][2]string
		system  [][2]string
		want    []string
	}{
		{
			name:    "dell keeps it on the manager, next to a fault list also typed SEL",
			manager: [][2]string{{"Sel", "SEL"}, {"FaultList", "SEL"}, {"Lclog", "Event"}},
			want:    []string{"Sel"},
		},
		{
			name:    "a copy under systems is not read twice",
			manager: [][2]string{{"Sel", "SEL"}},
			system:  [][2]string{{"Sel", "SEL"}},
			want:    []string{managerPath + "/LogServices/Sel"},
		},
		{
			name:    "a typed sel under systems is used when the managers type none",
			manager: [][2]string{{"IEL", "Event"}},
			system:  [][2]string{{"StandardLog", "Event"}, {"SEL", "SEL"}},
			want:    []string{"SEL"},
		},
		{
			name:    "a typed sel under systems wins over a manager log known only by its id",
			manager: [][2]string{{"Log1", "Multiple"}},
			system:  [][2]string{{"Log1", "SEL"}},
			want:    []string{systemPath + "/LogServices/Log1"},
		},
		{
			name:    "an untyped sel is found by its id beside other logs",
			manager: [][2]string{{"Sel", ""}, {"Lclog", ""}},
			want:    []string{"Sel"},
		},
		{
			name:    "supermicro's log1 on the manager is used when nothing is typed sel",
			manager: [][2]string{{"Log1", "Multiple"}, {"Audit", "Event"}},
			want:    []string{"Log1"},
		},
		{
			name:    "a lone untyped service has nothing to be confused with",
			manager: [][2]string{{"Whatever", ""}},
			want:    []string{"Whatever"},
		},
		{
			name:    "a lone log typed as something else is not the event log",
			manager: [][2]string{{"IEL", "Event"}},
			system:  [][2]string{{"IML", "Oem"}},
			want:    nil,
		},
		{
			name:    "a log that is none of these is not guessed at",
			manager: [][2]string{{"Audit", "Event"}, {"Other", "Event"}},
			system:  [][2]string{{"Debug", ""}},
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := logVendorFixture(tt.manager, tt.system).serve(t)

			c, err := testService(nil).connect(context.Background(), Target{Name: "n", Addr: srv.URL})
			require.NoError(t, err)
			defer c.Close()

			services, err := selLogServices(c)
			require.NoError(t, err)

			// a want starting with a slash is checking where the log was found, not just which one
			got := make([]string, 0, len(services))
			for _, l := range services {
				if len(tt.want) > 0 && strings.HasPrefix(tt.want[0], "/") {
					got = append(got, l.ODataID)
				} else {
					got = append(got, l.ID)
				}
			}
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestSELStopsAtTheManagersWhenTheyHaveIt(t *testing.T) {
	f := logVendorFixture([][2]string{{"Sel", "SEL"}}, nil)
	srv := f.serve(t)

	results := testService(nil).SEL(context.Background(), []Target{{Name: "n", Addr: srv.URL}}, SELRequest{})

	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Zero(t, f.hits["/redfish/v1/Systems"], "the system walk costs round trips and is only needed when the managers have nothing")
}

func TestSELReadsFromSystems(t *testing.T) {
	srv := logVendorFixture([][2]string{{"IEL", "Event"}}, [][2]string{{"SEL", "SEL"}}).serve(t)

	results := testService(nil).SEL(context.Background(), []Target{{Name: "n", Addr: srv.URL}}, SELRequest{})

	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	require.Len(t, results[0].Data.Entries, 1)
	assert.Equal(t, "SEL", results[0].Data.Entries[0].ID)
}

func TestClearSELPartialFailure(t *testing.T) {
	f := logVendorFixture([][2]string{{"Sel", "SEL"}, {"Sel2", "SEL"}}, nil)
	var cleared []string
	f.handler = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method != http.MethodPost {
			return false
		}
		cleared = append(cleared, r.URL.Path)
		if strings.Contains(r.URL.Path, "/Sel2/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(dellGeneralError))
			return true
		}
		w.WriteHeader(http.StatusOK)
		return true
	}
	srv := f.serve(t)

	results := testService(nil).ClearSEL(context.Background(), []Target{{Name: "n", Addr: srv.URL}})

	require.Len(t, results, 1)
	assert.False(t, results[0].Success, "a log that refused means the node did not fully clear")
	assert.Contains(t, results[0].Message, "cleared 1 of 2 log service(s), failed on Sel2")
	assert.Len(t, cleared, 2)
}

// TestClearSELSkipsTheFaultList pins that a Dell node with the usual pair of SEL typed logs is not reported as failed, since the fault list always refuses
func TestClearSELSkipsTheFaultList(t *testing.T) {
	f := logVendorFixture([][2]string{{"Sel", "SEL"}, {"FaultList", "SEL"}}, nil)
	f.handler = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method != http.MethodPost {
			return false
		}
		if strings.Contains(r.URL.Path, "/FaultList/") {
			w.WriteHeader(http.StatusBadRequest)
			return true
		}
		w.WriteHeader(http.StatusOK)
		return true
	}
	srv := f.serve(t)

	results := testService(nil).ClearSEL(context.Background(), []Target{{Name: "n", Addr: srv.URL}})

	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "cleared 1 log service(s)", results[0].Message)
}
