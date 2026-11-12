// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package beacon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSystem walks through states, one per poll, and stays on the last
type fakeSystem struct {
	states []string
	jobs   []string
	failed []FailedUnit
}

func (f *fakeSystem) State(context.Context) (string, error) {
	s := f.states[0]
	if len(f.states) > 1 {
		f.states = f.states[1:]
	}
	return s, nil
}

func (f *fakeSystem) FailedUnits(context.Context) ([]FailedUnit, error) { return f.failed, nil }
func (f *fakeSystem) Jobs(context.Context) ([]string, error)            { return f.jobs, nil }
func (f *fakeSystem) Close()                                            {}

type recorder struct {
	reports []Report
	err     error
}

func (r *recorder) Post(_ context.Context, rep Report) error {
	r.reports = append(r.reports, rep)
	return r.err
}

func (r *recorder) PostRetry(ctx context.Context, rep Report, _ time.Duration) error {
	return r.Post(ctx, rep)
}

func opts() WatchOptions {
	return WatchOptions{Interval: time.Millisecond, Heartbeat: time.Hour, Timeout: time.Hour, BootID: "b1"}
}

func TestWatch(t *testing.T) {
	t.Run("clean boot", func(t *testing.T) {
		sys := &fakeSystem{states: []string{"starting", "starting", "running"}, jobs: []string{"a.service"}}
		rec := &recorder{}
		o := opts()
		o.Complete = true

		require.NoError(t, Watch(context.Background(), sys, rec, o))
		require.Len(t, rec.reports, 2, "an unchanged booting state is only sent once")
		assert.Equal(t, Report{State: StateBooting, QueuedJobs: 1, BootID: "b1", Message: "system is starting"}, rec.reports[0])
		assert.Equal(t, Report{State: StateCompleted, BootID: "b1", Complete: true, Message: "system is running"}, rec.reports[1])
	})

	t.Run("degraded", func(t *testing.T) {
		sys := &fakeSystem{states: []string{"degraded"}, failed: []FailedUnit{{Name: "b.service"}, {Name: "a.mount"}}}
		rec := &recorder{}
		o := opts()
		o.Complete = true

		assert.ErrorIs(t, Watch(context.Background(), sys, rec, o), ErrBootFailed)
		require.Len(t, rec.reports, 1)
		r := rec.reports[0]
		assert.Equal(t, StateFailed, r.State)
		assert.False(t, r.Complete, "a failed boot never unprovisions")
		assert.Equal(t, "2 failed: a.mount, b.service", r.Message)
		assert.Len(t, r.FailedUnits, 2)
	})

	t.Run("timeout", func(t *testing.T) {
		sys := &fakeSystem{states: []string{"starting"}, jobs: []string{"slow.service"}}
		rec := &recorder{}
		o := opts()
		o.Timeout = 5 * time.Millisecond

		assert.ErrorIs(t, Watch(context.Background(), sys, rec, o), ErrBootFailed)
		last := rec.reports[len(rec.reports)-1]
		assert.Equal(t, StateFailed, last.State)
		assert.Equal(t, "boot did not finish within 5ms, still waiting on slow.service", last.Message)
	})

	t.Run("rejected stops the watch", func(t *testing.T) {
		sys := &fakeSystem{states: []string{"starting"}}
		rec := &recorder{err: ErrRejected}

		assert.ErrorIs(t, Watch(context.Background(), sys, rec, opts()), ErrRejected)
		assert.Len(t, rec.reports, 1)
	})
}

func TestClientPost(t *testing.T) {
	var got Report
	code := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(code)
		_, _ = w.Write([]byte(`{"message":"invalid token"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, false)
	require.NoError(t, c.Post(context.Background(), Report{State: StateCompleted, BootID: "b"}))
	assert.Equal(t, Report{State: StateCompleted, BootID: "b"}, got)

	code = http.StatusBadRequest
	err := c.PostRetry(context.Background(), Report{State: StateCompleted}, time.Minute)
	assert.ErrorIs(t, err, ErrRejected, "a 4xx is not retried")
	assert.True(t, strings.Contains(err.Error(), "invalid token"))

	code = http.StatusInternalServerError
	err = c.PostRetry(context.Background(), Report{State: StateCompleted}, 50*time.Millisecond)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrRejected)
}

func TestListNames(t *testing.T) {
	names := make([]string, 12)
	for i := range names {
		names[i] = string(rune('a' + i))
	}
	assert.Equal(t, "a, b, c, d, e, f, g, h, i, j and 2 more", listNames(names))
}

func TestWatchSendsFacts(t *testing.T) {
	sys := &fakeSystem{states: []string{"starting", "starting", "starting", "running"}, jobs: []string{"a"}}
	rec := &recorder{}
	o := opts()
	o.Heartbeat = 0
	calls := 0
	o.Facts = func(context.Context) map[string]string {
		calls++
		return map[string]string{"os.image_version": "2026.11.3"}
	}

	require.NoError(t, Watch(context.Background(), sys, rec, o))
	require.GreaterOrEqual(t, len(rec.reports), 3)
	assert.NotNil(t, rec.reports[0].Facts, "the first progress report carries the facts")
	assert.Nil(t, rec.reports[1].Facts, "later progress reports don't repeat them")
	assert.Equal(t, "2026.11.3", rec.reports[len(rec.reports)-1].Facts["os.image_version"], "the final report collects them again")
	assert.Equal(t, 2, calls)
}

func (f *fakeSystem) BootTimes(context.Context) (map[string]uint64, error) { return nil, nil }
