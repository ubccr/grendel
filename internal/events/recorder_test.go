// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package events

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/store/sqlstore"
	"github.com/ubccr/grendel/pkg/model"
)

func newRecorder(t *testing.T) *Recorder {
	t.Helper()
	db, err := sqlstore.New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return New(db)
}

func TestRecord(t *testing.T) {
	r := newRecorder(t)

	e := r.Record(model.Event{Kind: model.EventKindAudit, Action: "node.save", Summary: "Saved nodes", Subject: "cpn-[01-02]", Actor: "admin", Severity: model.SeveritySuccess})
	got, err := r.DB.LoadEvent(e.ID)
	require.NoError(t, err)
	assert.Equal(t, model.EventStateDone, got.State)
	assert.Equal(t, model.SeveritySuccess, got.Severity)
	assert.Equal(t, "admin", got.Actor)
	require.NotNil(t, got.EndedAt)

	// with no actor it is the system's
	e = r.Record(model.Event{Kind: model.EventKindAudit, Action: "db.restore", Summary: "Restored the database", Severity: model.SeveritySuccess})
	got, err = r.DB.LoadEvent(e.ID)
	require.NoError(t, err)
	assert.Equal(t, model.EventActorSystem, got.Actor)
}

func TestTrackedEvent(t *testing.T) {
	r := newRecorder(t)
	load := func(e *model.Event) *model.Event {
		t.Helper()
		got, err := r.DB.LoadEvent(e.ID)
		require.NoError(t, err)
		return got
	}

	e := r.Open(model.Event{Kind: model.EventKindFirmware, Action: "bmc.firmware.install", Summary: "bmc firmware install", Nodes: model.EventNodeList{
		{Node: "n1", State: model.EventNodePending},
		{Node: "n2", State: model.EventNodePending},
	}})
	assert.Nil(t, e.Nodes, "the nodes are written, and not kept on the event returned")
	got := load(e)
	assert.Equal(t, model.EventStateRunning, got.State)
	assert.Equal(t, model.SeverityInfo, got.Severity)
	assert.Equal(t, 2, got.Counts.Pending)
	assert.False(t, got.Nodes[0].UpdatedAt.IsZero(), "a node written without a time gets now")

	// writing nodes settles the event, which stays open while any node is not finished
	r.SetNodes(e, model.EventNode{Node: "n1", State: model.EventNodeSucceeded})
	assert.Equal(t, model.EventStateRunning, load(e).State)

	r.SetNodes(e, model.EventNode{Node: "n2", State: model.EventNodeStalled})
	got = load(e)
	assert.Equal(t, model.EventStateDone, got.State)
	assert.Equal(t, model.SeverityWarning, got.Severity, "one node never finished")
	require.NotNil(t, got.EndedAt)
	ended := *got.EndedAt

	// a node reporting after the event ended changes how it went, and not when it ended
	r.SetNodes(e, model.EventNode{Node: "n2", State: model.EventNodeSucceeded})
	got = load(e)
	assert.Equal(t, model.SeveritySuccess, got.Severity)
	assert.Equal(t, model.EventStateDone, got.State)
	assert.True(t, got.EndedAt.Equal(ended))
}

func TestTidy(t *testing.T) {
	r := newRecorder(t)
	viper.Set("events.retention", "24h")
	t.Cleanup(func() { viper.Set("events.retention", "2160h") })
	now := time.Now().UTC()

	// a provisioning whose last node went quiet is closed once that node counts as stalled
	quiet := r.Open(model.Event{Kind: model.EventKindProvision, Action: "provision", Summary: "Provisioning", Nodes: model.EventNodeList{
		{Node: "n1", State: model.EventNodeSucceeded},
		{Node: "n2", State: model.EventNodeRunning, Stage: "config", UpdatedAt: now.Add(-ProvisionStalledAfter - time.Second)},
	}})
	// one with a node heard from lately is left running
	busy := r.Open(model.Event{Kind: model.EventKindFirmware, Action: "bmc.firmware.install", Summary: "bmc firmware install", Nodes: model.EventNodeList{
		{Node: "n3", State: model.EventNodeRunning, UpdatedAt: now.Add(-time.Minute)},
	}})
	// a finished event past the retention is deleted
	old := r.Record(model.Event{Kind: model.EventKindAudit, Action: "node.save", Summary: "Saved nodes", Severity: model.SeveritySuccess})
	old.StartedAt = now.Add(-48 * time.Hour)
	require.NoError(t, r.DB.StoreEvent(old))
	// StoreEvent keeps when an event started, so the old one is written again under a new id
	old.ID = "0" + old.ID[1:]
	require.NoError(t, r.DB.StoreEvent(old))

	r.Tidy(now)

	got, err := r.DB.LoadEvent(quiet.ID)
	require.NoError(t, err)
	assert.Equal(t, model.EventStateDone, got.State)
	assert.Equal(t, model.SeverityWarning, got.Severity)
	assert.Equal(t, 1, got.Counts.Stalled)

	got, err = r.DB.LoadEvent(busy.ID)
	require.NoError(t, err)
	assert.Equal(t, model.EventStateRunning, got.State)

	_, err = r.DB.LoadEvent(old.ID)
	assert.Error(t, err, "the old event is gone")
}
