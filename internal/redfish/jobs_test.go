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

	"github.com/stmcginnis/gofish/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const jobServicePath = "/redfish/v1/JobService"

// jobFixture serves a queue of jobs and records every DELETE as it arrived on the wire. A DELETE of a job in refuse answers 400
func jobFixture(ids []string, refuse ...string) (*fixture, *[]string) {
	f := dellFixture()
	f.tree[jobServicePath] = fmt.Sprintf(`{
	  "@odata.id": %q,
	  "@odata.type": "#JobService.v1_0_0.JobService",
	  "Id": "JobService",
	  "Name": "Job Service",
	  "Jobs": {"@odata.id": %q}
	}`, jobServicePath, jobServicePath+"/Jobs")

	members := make([]string, 0, len(ids))
	for _, id := range ids {
		uri := jobServicePath + "/Jobs/" + id
		members = append(members, fmt.Sprintf(`{"@odata.id": %q}`, uri))
		f.tree[uri] = fmt.Sprintf(`{
		  "@odata.id": %q,
		  "@odata.type": "#Job.v1_0_0.Job",
		  "Id": %q,
		  "Name": "Job",
		  "JobState": "Scheduled"
		}`, uri, id)
	}
	f.tree[jobServicePath+"/Jobs"] = fmt.Sprintf(`{
	  "@odata.id": %q,
	  "Members@odata.count": %d,
	  "Members": [%s]
	}`, jobServicePath+"/Jobs", len(ids), strings.Join(members, ","))

	deleted := &[]string{}
	f.handler = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method != http.MethodDelete {
			return false
		}
		*deleted = append(*deleted, r.URL.EscapedPath())

		for _, id := range refuse {
			if strings.HasSuffix(r.URL.Path, "/"+id) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(dellGeneralError))
				return true
			}
		}
		w.WriteHeader(http.StatusOK)
		return true
	}

	return f, deleted
}

func clearJobs(t *testing.T, f *fixture, req ClearJobsRequest) NodeResult[Ack] {
	t.Helper()

	results := testService(nil).ClearJobs(context.Background(), []Target{{Name: "cpn-a01-01", Addr: f.serve(t).URL}}, req)
	require.Len(t, results, 1)
	return results[0]
}

func TestClearAllJobs(t *testing.T) {
	f, deleted := jobFixture([]string{"JID_1", "JID_2"})

	r := clearJobs(t, f, ClearJobsRequest{All: true})

	assert.True(t, r.Success, "message was: %s", r.Message)
	assert.Equal(t, "cleared 2 job(s)", r.Message)
	assert.ElementsMatch(t, []string{jobServicePath + "/Jobs/JID_1", jobServicePath + "/Jobs/JID_2"}, *deleted)
}

func TestClearAllJobsEmptyQueue(t *testing.T) {
	f, deleted := jobFixture(nil)

	r := clearJobs(t, f, ClearJobsRequest{All: true})

	assert.True(t, r.Success)
	assert.Equal(t, "no jobs to clear", r.Message)
	assert.Empty(t, *deleted)
}

// TestClearJobsPartialFailure checks a stuck job neither stops the rest nor hides how far the clear got
func TestClearJobsPartialFailure(t *testing.T) {
	f, deleted := jobFixture([]string{"JID_1", "JID_2", "JID_3"}, "JID_1")

	r := clearJobs(t, f, ClearJobsRequest{IDs: []string{"JID_1", "JID_2", "JID_3"}})

	assert.False(t, r.Success)
	assert.Contains(t, r.Message, "cleared 2 of 3 job(s), failed on JID_1")
	assert.Contains(t, r.Message, "the JobID entered is invalid", "the reason has to survive")
	assert.Len(t, *deleted, 3, "every job has to be attempted")
	assert.Nil(t, r.Data)
}

// TestClearJobsEscapesID is why the id is escaped: gofish does not clean the path, so this would otherwise be a DELETE of a BMC account, sent with admin credentials
func TestClearJobsEscapesID(t *testing.T) {
	f, deleted := jobFixture(nil)

	clearJobs(t, f, ClearJobsRequest{IDs: []string{"../../AccountService/Accounts/3"}})

	require.Len(t, *deleted, 1)
	assert.Equal(t, jobServicePath+"/Jobs/..%2F..%2FAccountService%2FAccounts%2F3", (*deleted)[0])
}

func TestClearJobsRejectsEmptyID(t *testing.T) {
	f, deleted := jobFixture(nil)

	r := clearJobs(t, f, ClearJobsRequest{IDs: []string{"JID_1", ""}})

	assert.False(t, r.Success)
	assert.Empty(t, *deleted, "an empty id would target the Jobs collection itself")
}

// TestClearJobsTrailingSlash covers a service whose JobService @odata.id ends in a slash, which must not produce //Jobs
func TestClearJobsTrailingSlash(t *testing.T) {
	f, deleted := jobFixture(nil)
	f.tree[jobServicePath] = strings.Replace(f.tree[jobServicePath], fmt.Sprintf(`"@odata.id": %q`, jobServicePath), fmt.Sprintf(`"@odata.id": %q`, jobServicePath+"/"), 1)

	r := clearJobs(t, f, ClearJobsRequest{IDs: []string{"JID_1"}})

	assert.True(t, r.Success, "message was: %s", r.Message)
	assert.Equal(t, []string{jobServicePath + "/Jobs/JID_1"}, *deleted)
}

func TestSortJobsNewestFirst(t *testing.T) {
	jobs := []*schemas.Job{
		{Entity: schemas.Entity{ID: "JID_9"}, StartTime: "TIME_NOW"},
		{Entity: schemas.Entity{ID: "JID_10"}, StartTime: "TIME_NA"},
		{Entity: schemas.Entity{ID: "JID_1"}, StartTime: "2026-01-02T00:00:00Z"},
		{Entity: schemas.Entity{ID: "JID_2"}, CreationTime: "2026-03-01T00:00:00-05:00", StartTime: "2026-01-01T00:00:00Z"},
		{Entity: schemas.Entity{ID: "JID_3"}, EndTime: "2026-02-01T00:00:00Z"},
	}

	sortJobsNewestFirst(jobs)

	ids := make([]string, 0, len(jobs))
	for _, j := range jobs {
		ids = append(ids, j.ID)
	}
	assert.Equal(t, []string{"JID_2", "JID_3", "JID_1", "JID_10", "JID_9"}, ids)
}

func TestJobsNewestFirst(t *testing.T) {
	f, _ := jobFixture([]string{"JID_100", "JID_200", "JID_300"})

	results := testService(nil).Jobs(context.Background(), []Target{{Name: "cpn-a01-01", Addr: f.serve(t).URL}})
	require.Len(t, results, 1)
	require.True(t, results[0].Success, results[0].Message)

	var ids []string
	for _, j := range results[0].Data.Jobs {
		ids = append(ids, j.ID)
	}
	assert.Equal(t, []string{"JID_300", "JID_200", "JID_100"}, ids)
}
