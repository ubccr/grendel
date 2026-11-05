// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockRedfish starts a fake BMC serving the given path to JSON body routes and returns a client connected to it
func newMockRedfish(t *testing.T, routes map[string]string) *Redfish {
	t.Helper()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, ok := routes[req.URL.Path]
		if !ok {
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	r, err := NewRedfishClient(srv.Listener.Addr().String(), "", "", true)
	require.NoError(t, err)

	return r
}

func TestRedfishNoJobService(t *testing.T) {
	r := newMockRedfish(t, map[string]string{
		"/redfish/v1/": `{"@odata.id": "/redfish/v1/", "Id": "RootService", "RedfishVersion": "1.6.0"}`,
	})

	_, err := r.GetJobs()
	assert.ErrorIs(t, err, errNoJobService)

	_, err = r.GetJobInfo("JID_1")
	assert.ErrorIs(t, err, errNoJobService)

	_, err = r.BmcGetJob("JID_1")
	assert.ErrorIs(t, err, errNoJobService)

	err = r.ClearJobs([]string{"JID_CLEARALL"})
	assert.ErrorIs(t, err, errNoJobService)
}

func TestRedfishGetJobs(t *testing.T) {
	r := newMockRedfish(t, map[string]string{
		"/redfish/v1/":                      `{"@odata.id": "/redfish/v1/", "Id": "RootService", "RedfishVersion": "1.6.0", "JobService": {"@odata.id": "/redfish/v1/JobService"}}`,
		"/redfish/v1/JobService":            `{"@odata.id": "/redfish/v1/JobService", "Id": "JobService", "Jobs": {"@odata.id": "/redfish/v1/JobService/Jobs"}}`,
		"/redfish/v1/JobService/Jobs":       `{"@odata.id": "/redfish/v1/JobService/Jobs", "Members": [{"@odata.id": "/redfish/v1/JobService/Jobs/JID_1"}], "Members@odata.count": 1}`,
		"/redfish/v1/JobService/Jobs/JID_1": `{"@odata.id": "/redfish/v1/JobService/Jobs/JID_1", "Id": "JID_1", "Name": "test job"}`,
	})

	jobs, err := r.GetJobs()
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, "JID_1", jobs[0].ID)
}

func TestRedfish(t *testing.T) {
	endpoint := os.Getenv("GRENDEL_BMC_ENDPOINT")
	user := os.Getenv("GRENDEL_BMC_USER")
	pass := os.Getenv("GRENDEL_BMC_PASS")

	if endpoint == "" || user == "" || pass == "" {
		t.Skip("Skipping BMC test. Missing env vars")
	}

	r, err := NewRedfishClient(endpoint, user, pass, true)
	assert.Nil(t, err)
	defer r.client.Logout()

	system, err := r.GetSystem()
	assert.Nil(t, err)
	assert.Greater(t, len(system.BIOSVersion), 0)
}
