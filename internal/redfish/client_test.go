// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireBasicAuth answers everything past the service root with a 401 unless the password matches, which is how a real BMC behaves: the root is public, so connecting proves nothing about the credentials
func requireBasicAuth(f *fixture, password string) {
	f.handler = func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path == "/redfish/v1/" {
			return false
		}
		if _, got, ok := r.BasicAuth(); ok && got == password {
			return false
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(dellAuthError))
		return true
	}
}

func TestBasicAuthFallsBackToTheNextCredential(t *testing.T) {
	f := dellFixture()
	requireBasicAuth(f, "calvin")
	srv := f.serve(t)

	svc := testService(func(c *Config) {
		c.Credentials = []Credential{
			{User: "grendel", Password: "wrong"},
			{User: "root", Password: "calvin"},
		}
	})

	results := svc.Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})

	require.Len(t, results, 1)
	assert.True(t, results[0].Success, "message was: %s", results[0].Message)
}

func TestBasicAuthAllCredentialsRejected(t *testing.T) {
	f := dellFixture()
	requireBasicAuth(f, "something else")
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
	assert.NotEmpty(t, results[0].Message)
}

// TestBasicAuthSingleCredentialSkipsTheCheck keeps basic auth cheap for the usual config, where there is nothing to fall back to
func TestBasicAuthSingleCredentialSkipsTheCheck(t *testing.T) {
	f := dellFixture()
	srv := f.serve(t)

	results := testService(nil).Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})

	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, 1, f.hits["/redfish/v1/Systems"], "only the operation itself should read the collection")
}

func TestCollectionMemberRejectedReportsTheBMCsReason(t *testing.T) {
	f := dellFixture()
	f.status["/redfish/v1/Systems/System.Embedded.1"] = http.StatusUnauthorized
	f.tree["/redfish/v1/Systems/System.Embedded.1"] = dellAuthError
	srv := f.serve(t)

	results := testService(nil).Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})

	require.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.Contains(t, results[0].Message, "invalid username")
}

func TestCollectionMemberTimeoutFailsTheNode(t *testing.T) {
	f := dellFixture()
	f.delay["/redfish/v1/Systems/System.Embedded.1"] = 2 * time.Second
	srv := f.serve(t)

	svc := testService(func(c *Config) { c.NodeTimeout = 100 * time.Millisecond })

	results := svc.Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})

	require.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.Contains(t, results[0].Message, "deadline")
}

func TestNoBMCAddressIsAFailedRow(t *testing.T) {
	results := testService(nil).Status(context.Background(), []Target{{Name: "cpn-a01-04"}})

	require.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.Contains(t, results[0].Message, "no usable ip address")
}

const sessionsPath = "/redfish/v1/SessionService/Sessions"

// sessionFixture is a service that does session login, with logout optionally stuck for the given time
func sessionFixture(deleteDelay time.Duration) *fixture {
	f := dellFixture()
	f.tree["/redfish/v1/"] = strings.Replace(serviceRootDell, `"Id": "RootService",`,
		`"Id": "RootService", "Links": {"Sessions": {"@odata.id": "`+sessionsPath+`"}},`, 1)
	f.handler = func(w http.ResponseWriter, r *http.Request) bool {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == sessionsPath:
			w.Header().Set("X-Auth-Token", "token")
			w.Header().Set("Location", sessionsPath+"/1")
			w.WriteHeader(http.StatusCreated)
			return true
		case r.Method == http.MethodDelete && r.URL.Path == sessionsPath+"/1":
			select {
			case <-time.After(deleteDelay):
			case <-r.Context().Done():
				return true
			}
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		return false
	}
	return f
}

func warnings(hook *logrustest.Hook) []string {
	var out []string
	for _, e := range hook.AllEntries() {
		if e.Level == logrus.WarnLevel {
			out = append(out, e.Message)
		}
	}
	return out
}

func TestCloseEndsTheSession(t *testing.T) {
	f := sessionFixture(0)
	srv := f.serve(t)

	hook := logrustest.NewLocal(log.Logger)
	defer hook.Reset()

	svc := testService(func(c *Config) { c.BasicAuth = false })

	results := svc.Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})

	require.True(t, results[0].Success, "message was: %s", results[0].Message)
	assert.Equal(t, 1, f.hits[sessionsPath+"/1"], "the session must be deleted")
	assert.Empty(t, warnings(hook))
}

// TestCloseIsBoundedAndLogged is what keeps a wedged BMC from holding its fan-out slot for a whole client timeout, and from leaking a session nobody hears about
func TestCloseIsBoundedAndLogged(t *testing.T) {
	previous := logoutTimeout
	logoutTimeout = 50 * time.Millisecond
	t.Cleanup(func() { logoutTimeout = previous })

	srv := sessionFixture(5 * time.Second).serve(t)

	hook := logrustest.NewLocal(log.Logger)
	defer hook.Reset()

	svc := testService(func(c *Config) { c.BasicAuth = false })

	begin := time.Now()
	results := svc.Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})

	require.True(t, results[0].Success, "a stuck logout must not cost the result, message was: %s", results[0].Message)
	assert.Less(t, time.Since(begin), 2*time.Second)

	got := warnings(hook)
	require.Len(t, got, 1)
	assert.Contains(t, got[0], "cpn-a01-01")
}

// TestCloseAfterNodeDeadline checks the logout still goes out once the context the connection was made with has expired
func TestCloseAfterNodeDeadline(t *testing.T) {
	f := sessionFixture(0)
	f.delay["/redfish/v1/Systems"] = 2 * time.Second
	srv := f.serve(t)

	svc := testService(func(c *Config) {
		c.BasicAuth = false
		c.NodeTimeout = 200 * time.Millisecond
	})

	results := svc.Status(context.Background(), []Target{{Name: "cpn-a01-01", Addr: srv.URL}})

	require.False(t, results[0].Success)
	assert.Equal(t, 1, f.hits[sessionsPath+"/1"], "the session must not be left open")
}
