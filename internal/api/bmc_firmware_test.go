// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"testing"

	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
)

func job(id string, state schemas.JobState, percent uint) *schemas.Job {
	j := &schemas.Job{JobState: state, PercentComplete: &percent}
	j.ID = id
	j.Name = "Firmware Update"
	return j
}

func TestFirmwareProgress(t *testing.T) {
	queued := model.EventNode{Node: "n1", State: model.EventNodeRunning, Stage: firmwareStageQueued, Ref: "JID_100"}
	old := job("JID_99", schemas.ExceptionJobState, 0)

	for name, c := range map[string]struct {
		row     model.EventNode
		jobs    []*schemas.Job
		state   model.EventNodeState
		stage   string
		changed bool
	}{
		"the stage is the state of the oldest job not done": {
			queued, []*schemas.Job{old, job("JID_102", schemas.PendingJobState, 0), job("JID_101", schemas.RunningJobState, 40), job("JID_100", schemas.CompletedJobState, 100)},
			model.EventNodeRunning, "running", true,
		},
		"a job from before the update is not its failure": {
			queued, []*schemas.Job{old, job("JID_100", schemas.NewJobState, 0)},
			model.EventNodeRunning, "new", true,
		},
		"nothing changed is not written": {
			model.EventNode{Node: "n1", State: model.EventNodeRunning, Stage: "running", Ref: "JID_100"},
			[]*schemas.Job{job("JID_100", schemas.RunningJobState, 60)},
			model.EventNodeRunning, "running", false,
		},
		"every job done is checked once more before it counts": {
			queued, []*schemas.Job{job("JID_100", schemas.CompletedJobState, 100), job("JID_101", schemas.RunningJobState, 100)},
			model.EventNodeRunning, firmwareStageVerifying, true,
		},
		"and counts when it still reads done": {
			model.EventNode{Node: "n1", State: model.EventNodeRunning, Stage: firmwareStageVerifying, Ref: "JID_100"},
			[]*schemas.Job{job("JID_100", schemas.CompletedJobState, 100)},
			model.EventNodeSucceeded, firmwareStageVerifying, true,
		},
		"a job added while verifying puts it back to running": {
			model.EventNode{Node: "n1", State: model.EventNodeRunning, Stage: firmwareStageVerifying, Ref: "JID_100"},
			[]*schemas.Job{job("JID_100", schemas.CompletedJobState, 100), job("JID_101", schemas.PendingJobState, 0)},
			model.EventNodeRunning, "pending", true,
		},
		"a failed job fails the node": {
			queued, []*schemas.Job{job("JID_100", schemas.CompletedJobState, 100), job("JID_101", schemas.ExceptionJobState, 30)},
			model.EventNodeFailed, "exception", true,
		},
		"a queue without the job leaves the node as it is": {
			queued, []*schemas.Job{old},
			model.EventNodeRunning, firmwareStageQueued, false,
		},
	} {
		got, changed := firmwareProgress(c.row, c.jobs)
		if changed != c.changed || got.State != c.state || got.Stage != c.stage {
			t.Errorf("%s: got %s at %q changed=%t, want %s at %q changed=%t", name, got.State, got.Stage, changed, c.state, c.stage, c.changed)
		}
	}
}

// TestFirmwareInstallLeavesTheEventOpen is what lets the update be followed after the request that queued it has returned
func TestFirmwareInstallLeavesTheEventOpen(t *testing.T) {
	h := newTestHandler(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	if err := h.DB.StoreHost(host); err != nil {
		t.Fatal(err)
	}

	track := &bmcTrack[redfish.Ack]{kind: model.EventKindFirmware, action: "bmc.firmware.install", node: firmwareQueued}
	_, err := runBmc(h, newFakeBmcContext(map[string]string{"nodeset": host.Name}), "bmc firmware install from 10.0.0.1", track,
		func(context.Context, []redfish.Target) []redfish.NodeResult[redfish.Ack] {
			return []redfish.NodeResult[redfish.Ack]{
				{Node: host.Name, Success: true, Data: &redfish.Ack{Job: "JID_7"}},
				{Node: "cpn-9", Message: "connection refused"},
			}
		},
		func(*redfish.Ack) *model.BmcAck { return nil },
	)
	if err != nil {
		t.Fatal(err)
	}

	events := listEvents(t, h)
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	event, err := h.DB.LoadEvent(events[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != model.EventKindFirmware || event.State != model.EventStateRunning {
		t.Errorf("event is %s and %s, want a running firmware event", event.Kind, event.State)
	}
	if want := (model.EventCounts{Total: 2, Running: 1, Failed: 1}); event.Counts != want {
		t.Errorf("counts = %+v, want %+v", event.Counts, want)
	}
	for _, n := range event.Nodes {
		if n.Node == host.Name && (n.Stage != firmwareStageQueued || n.Ref != "JID_7") {
			t.Errorf("queued node is at %q with %q, want the job id to follow", n.Stage, n.Ref)
		}
	}
}
