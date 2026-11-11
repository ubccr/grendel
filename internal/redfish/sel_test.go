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

// selFixture builds a manager with one SEL log service whose entries are returned in the order given, which is how a real service behaves: the collection order is not the timestamp order
func selFixture(entries []selEntry) *fixture {
	f := dellFixture()
	f.tree["/redfish/v1/Managers"] = `{
	  "@odata.id": "/redfish/v1/Managers",
	  "Members@odata.count": 1,
	  "Members": [{"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1"}]
	}`
	f.tree["/redfish/v1/Managers/iDRAC.Embedded.1"] = `{
	  "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1",
	  "@odata.type": "#Manager.v1_5_0.Manager",
	  "Id": "iDRAC.Embedded.1",
	  "Name": "Manager",
	  "LogServices": {"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices"}
	}`
	f.tree["/redfish/v1/Managers/iDRAC.Embedded.1/LogServices"] = `{
	  "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices",
	  "Members@odata.count": 1,
	  "Members": [{"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel"}]
	}`
	f.tree["/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel"] = `{
	  "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel",
	  "@odata.type": "#LogService.v1_1_0.LogService",
	  "Id": "Sel",
	  "Name": "System Event Log",
	  "LogEntryType": "SEL",
	  "Entries": {"@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"}
	}`

	members := make([]string, 0, len(entries))
	for _, e := range entries {
		uri := fmt.Sprintf("/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries/%s", e.id)
		members = append(members, fmt.Sprintf(`{"@odata.id": %q}`, uri))
		f.tree[uri] = fmt.Sprintf(`{
		  "@odata.id": %q,
		  "@odata.type": "#LogEntry.v1_4_0.LogEntry",
		  "Id": %q,
		  "Name": "Log Entry %s",
		  "Created": %q,
		  "Severity": %q,
		  "EntryType": "SEL",
		  "Message": %q
		}`, uri, e.id, e.id, e.created, e.severity, e.message)
	}

	f.tree["/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"] = fmt.Sprintf(`{
	  "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries",
	  "Members@odata.count": %d,
	  "Members": [%s]
	}`, len(entries), strings.Join(members, ","))

	return f
}

type selEntry struct {
	id, created, severity, message string
}

// the ids are deliberately out of order, and so is the collection. A real iDRAC returned 11, 10, 13, 12, 9 ... which is why taking entries off either end of the collection picked the wrong ones
var scrambledSEL = []selEntry{
	{"11", "2026-02-23T15:44:58-06:00", "OK", "eleventh"},
	{"10", "2026-02-23T15:39:34-06:00", "OK", "tenth"},
	{"13", "2026-02-23T15:52:27-06:00", "Critical", "thirteenth"},
	{"12", "2026-02-23T15:45:06-06:00", "Warning", "twelfth"},
	{"9", "2024-06-28T15:06:16-05:00", "OK", "ninth"},
	{"1", "2023-09-14T15:21:23-05:00", "OK", "first"},
}

func TestSELReturnsNewestFirst(t *testing.T) {
	srv := selFixture(scrambledSEL).serve(t)

	results := testService(nil).SEL(context.Background(),
		[]Target{{Name: "cpn-a01-01", Addr: srv.URL}}, SELRequest{Limit: 3})

	require.Len(t, results, 1)
	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	require.NotNil(t, results[0].Data)

	got := make([]string, 0, 3)
	for _, e := range results[0].Data.Entries {
		got = append(got, e.ID)
	}

	// the fixture ignores $top, as some BMCs do, so the whole log comes back and the limit is applied after sorting
	assert.Equal(t, []string{"13", "12", "11"}, got, "a BMC that ignores $top must still get the newest entries, not whichever end of the collection they arrived on")
}

// TestSELSendsTop covers a BMC that honours $top: the limit is asked for rather than the whole log being read, and what comes back is the first members in the BMC's own order, sorted newest first
func TestSELSendsTop(t *testing.T) {
	f := selFixture(scrambledSEL)
	var query string
	f.handler = func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path != "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries" {
			return false
		}
		query = r.URL.RawQuery
		// a BMC that honours $top=2 lists only its first two members
		top := selFixture(scrambledSEL[:2]).tree[r.URL.Path]
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(top))
		return true
	}
	srv := f.serve(t)

	results := testService(nil).SEL(context.Background(),
		[]Target{{Name: "cpn-a01-01", Addr: srv.URL}}, SELRequest{Limit: 2})

	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, "$top=2", query)

	got := make([]string, 0, 2)
	for _, e := range results[0].Data.Entries {
		got = append(got, e.ID)
	}
	assert.Equal(t, []string{"11", "10"}, got, "the first two members the BMC listed, newest first")
}

func TestSELUnlimitedReturnsEverything(t *testing.T) {
	srv := selFixture(scrambledSEL).serve(t)

	results := testService(nil).SEL(context.Background(),
		[]Target{{Name: "cpn-a01-01", Addr: srv.URL}}, SELRequest{})

	require.NotNil(t, results[0].Data)
	assert.Len(t, results[0].Data.Entries, len(scrambledSEL))
	assert.Equal(t, "1", results[0].Data.Entries[len(scrambledSEL)-1].ID, "the oldest entry sorts last")
}

// TestSELNoLogService covers a bmc that answers but keeps no system event log
func TestSELNoLogService(t *testing.T) {
	f := dellFixture()
	f.tree["/redfish/v1/Managers"] = emptyCollection

	srv := f.serve(t)

	results := testService(nil).SEL(context.Background(),
		[]Target{{Name: "cpn-a01-01", Addr: srv.URL}}, SELRequest{})

	require.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.Contains(t, results[0].Message, "no system event log")
}
