// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/viper"

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

// dellJob is a job as an iDRAC lists it, with its own job state under Oem beside the Redfish one
func dellJob(id string, state schemas.JobState, dellState string) *schemas.Job {
	j := job(id, state, 0)
	j.OEM = json.RawMessage(fmt.Sprintf(`{"Dell":{"JobState":%q}}`, dellState))
	return j
}

func TestFirmwareProgress(t *testing.T) {
	queued := model.EventNode{Node: "n1", State: model.EventNodeRunning, Stage: firmwareStageQueued, Ref: "JID_100"}
	stalled := model.EventNode{Node: "n1", State: model.EventNodeStalled, Stage: "running", Ref: "JID_100"}
	old := job("JID_99", schemas.ExceptionJobState, 0)

	for name, c := range map[string]struct {
		row     model.EventNode
		jobs    []*schemas.Job
		wait    bool
		state   model.EventNodeState
		stage   string
		changed bool
	}{
		"the stage is the state of the oldest job not done": {
			queued, []*schemas.Job{old, job("JID_102", schemas.PendingJobState, 0), job("JID_101", schemas.RunningJobState, 40), job("JID_100", schemas.CompletedJobState, 100)}, false,
			model.EventNodeRunning, "running", true,
		},
		"a job from before the update is not its failure": {
			queued, []*schemas.Job{old, job("JID_100", schemas.NewJobState, 0)}, false,
			model.EventNodeRunning, "new", true,
		},
		"nothing changed is not a change": {
			model.EventNode{Node: "n1", State: model.EventNodeRunning, Stage: "running", Ref: "JID_100"},
			[]*schemas.Job{job("JID_100", schemas.RunningJobState, 60)}, false,
			model.EventNodeRunning, "running", false,
		},
		"every job done is checked once more before it counts": {
			queued, []*schemas.Job{job("JID_100", schemas.CompletedJobState, 100), job("JID_101", schemas.RunningJobState, 100)}, false,
			model.EventNodeRunning, firmwareStageVerifying, true,
		},
		"and counts when it still reads done": {
			model.EventNode{Node: "n1", State: model.EventNodeRunning, Stage: firmwareStageVerifying, Ref: "JID_100"},
			[]*schemas.Job{job("JID_100", schemas.CompletedJobState, 100)}, false,
			model.EventNodeSucceeded, firmwareStageVerifying, true,
		},
		"a job added while verifying puts it back to running": {
			model.EventNode{Node: "n1", State: model.EventNodeRunning, Stage: firmwareStageVerifying, Ref: "JID_100"},
			[]*schemas.Job{job("JID_100", schemas.CompletedJobState, 100), job("JID_101", schemas.PendingJobState, 0)}, false,
			model.EventNodeRunning, "pending", true,
		},
		"a failed job fails the node": {
			queued, []*schemas.Job{job("JID_100", schemas.CompletedJobState, 100), job("JID_101", schemas.ExceptionJobState, 30)}, false,
			model.EventNodeFailed, "exception", true,
		},
		"a job the iDRAC says completed with errors fails the node, whatever the Redfish state says": {
			queued, []*schemas.Job{dellJob("JID_100", schemas.CompletedJobState, "CompletedWithErrors")}, false,
			model.EventNodeFailed, "completedwitherrors", true,
		},
		"a queue without the job stalls the node, as how it went cannot be known": {
			queued, []*schemas.Job{old}, false,
			model.EventNodeStalled, firmwareStageGone, true,
		},
		"and once stalled for that there is nothing new": {
			model.EventNode{Node: "n1", State: model.EventNodeStalled, Stage: firmwareStageGone, Ref: "JID_100", Message: "job JID_100 is no longer in the BMC's job queue"},
			[]*schemas.Job{old}, false,
			model.EventNodeStalled, firmwareStageGone, false,
		},
		"an update left for a reboot the install was asked not to do is done with": {
			queued, []*schemas.Job{job("JID_100", schemas.CompletedJobState, 100), dellJob("JID_101", schemas.PendingJobState, "Scheduled"), dellJob("JID_102", schemas.PendingJobState, "Scheduled")}, true,
			model.EventNodeSucceeded, firmwareStageScheduled, true,
		},
		"but not while any of it is still going": {
			queued, []*schemas.Job{dellJob("JID_100", schemas.RunningJobState, "Running"), dellJob("JID_101", schemas.PendingJobState, "Scheduled")}, true,
			model.EventNodeRunning, "running", true,
		},
		"and not when the install was let reboot, as the reboot is coming": {
			queued, []*schemas.Job{dellJob("JID_100", schemas.PendingJobState, "Scheduled")}, false,
			model.EventNodeRunning, "pending", true,
		},
		"a stalled node heard from again stays stalled while its update goes on": {
			stalled, []*schemas.Job{job("JID_100", schemas.CompletedJobState, 100), job("JID_101", schemas.RunningJobState, 50)}, false,
			model.EventNodeStalled, "running", false,
		},
		"and takes its outcome once it has one": {
			model.EventNode{Node: "n1", State: model.EventNodeStalled, Stage: firmwareStageVerifying, Ref: "JID_100"},
			[]*schemas.Job{job("JID_100", schemas.CompletedJobState, 100)}, false,
			model.EventNodeSucceeded, firmwareStageVerifying, true,
		},
	} {
		got, changed := firmwareProgress(c.row, c.jobs, c.wait)
		if changed != c.changed || got.State != c.state || got.Stage != c.stage {
			t.Errorf("%s: got %s at %q changed=%t, want %s at %q changed=%t", name, got.State, got.Stage, changed, c.state, c.stage, c.changed)
		}
	}
}

// fakeJobs answers job reads from a queue per node, and as unreachable for a node it has none for
type fakeJobs map[string][]*schemas.Job

func (f fakeJobs) read(_ context.Context, targets []redfish.Target) []redfish.NodeResult[redfish.JobList] {
	results := make([]redfish.NodeResult[redfish.JobList], 0, len(targets))
	for _, t := range targets {
		jobs, ok := f[t.Name]
		if !ok {
			results = append(results, redfish.NodeResult[redfish.JobList]{Node: t.Name, Message: "connection refused"})
			continue
		}
		results = append(results, redfish.NodeResult[redfish.JobList]{Node: t.Name, Success: true, Data: &redfish.JobList{Jobs: jobs}})
	}
	return results
}

// TestCheckFirmware follows an update across a BMC that goes quiet and comes back, which is what a BMC applying its own firmware does
func TestCheckFirmware(t *testing.T) {
	h := newTestHandler(t)
	names := []string{"fw-1", "fw-2"}
	for _, name := range names {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = name
		if err := h.DB.StoreHost(host); err != nil {
			t.Fatal(err)
		}
	}
	timeout := viper.GetDuration("bmc.firmware_timeout")

	event := h.Events.Open(model.Event{
		Kind:    model.EventKindFirmware,
		Action:  "bmc.firmware.install",
		Summary: "bmc firmware install",
		Detail:  map[string]string{firmwareDetailReboot: "true"},
		Nodes: model.EventNodeList{
			// fw-1 is already at the stage its BMC will report, so what it answers is no change
			{Node: "fw-1", State: model.EventNodeRunning, Stage: "running", Ref: "JID_1"},
			{Node: "fw-2", State: model.EventNodeRunning, Stage: firmwareStageQueued, Ref: "JID_1"},
		},
	})
	load := func() *model.Event {
		t.Helper()
		got, err := h.DB.LoadEvent(event.ID)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	node := func(e *model.Event, name string) model.EventNode {
		t.Helper()
		for _, n := range e.Nodes {
			if n.Node == name {
				return n
			}
		}
		t.Fatalf("event has no node %s", name)
		return model.EventNode{}
	}
	// a check made as if from later, with the nodes last heard from at the given time
	age := func(name string, heard time.Time) {
		t.Helper()
		n := node(load(), name)
		n.UpdatedAt = heard
		if err := h.DB.StoreEventNodes(event.ID, model.EventNodeList{n}, heard); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC()

	// fw-1 answers with its job still running, unchanged, and fw-2 does not answer
	running := fakeJobs{"fw-1": {job("JID_1", schemas.RunningJobState, 30)}}
	age("fw-1", now.Add(-timeout-time.Minute))
	age("fw-2", now.Add(-timeout-time.Minute))
	h.checkFirmware(context.Background(), now, running.read)
	if got := node(load(), "fw-1"); now.Sub(got.UpdatedAt) > time.Minute {
		t.Errorf("fw-1 answered and was last written %s, want it written as heard from now", got.UpdatedAt)
	}

	// so only the node that did not answer is stalled, though neither changed
	h.Events.Tidy(now)
	got := load()
	if s := node(got, "fw-1").State; s != model.EventNodeRunning {
		t.Errorf("fw-1 is %s, want running, its BMC answers", s)
	}
	if s := node(got, "fw-2").State; s != model.EventNodeStalled {
		t.Errorf("fw-2 is %s, want stalled, its BMC has been quiet past the timeout", s)
	}

	// fw-1 finishes, which ends the event as a warning with fw-2 stalled
	done := fakeJobs{"fw-1": {job("JID_1", schemas.CompletedJobState, 100)}}
	h.checkFirmware(context.Background(), now, done.read)
	h.checkFirmware(context.Background(), now, done.read)
	got = load()
	if got.State != model.EventStateDone || got.Severity != model.SeverityWarning {
		t.Fatalf("event is %s %s, want done as a warning", got.State, got.Severity)
	}

	// fw-2's BMC comes back with its update done, and the ended event takes the outcome
	back := fakeJobs{"fw-1": {job("JID_1", schemas.CompletedJobState, 100)}, "fw-2": {job("JID_1", schemas.CompletedJobState, 100)}}
	h.checkFirmware(context.Background(), now, back.read)
	h.checkFirmware(context.Background(), now, back.read)
	got = load()
	if s := node(got, "fw-2").State; s != model.EventNodeSucceeded {
		t.Errorf("fw-2 is %s, want succeeded once its BMC came back", s)
	}
	if got.Severity != model.SeveritySuccess {
		t.Errorf("event is %s, want success once every node succeeded", got.Severity)
	}

	// past firmwareFollowFor nothing is checked any more
	stalledAgain := fakeJobs{"fw-1": {}, "fw-2": {}}
	h.checkFirmware(context.Background(), now.Add(firmwareFollowFor+time.Hour), stalledAgain.read)
	if s := node(load(), "fw-2").State; s != model.EventNodeSucceeded {
		t.Errorf("fw-2 is %s, an event past firmwareFollowFor is left alone", s)
	}
}

// TestCheckFirmwareStallsAnUpdateThatNeverFinishes keeps a BMC that answers but never finishes from holding its event open for ever
func TestCheckFirmwareStallsAnUpdateThatNeverFinishes(t *testing.T) {
	h := newTestHandler(t)
	host := tests.HostFactory.MustCreate().(*model.Host)
	host.Name = "fw-1"
	if err := h.DB.StoreHost(host); err != nil {
		t.Fatal(err)
	}
	event := h.Events.Open(model.Event{
		Kind: model.EventKindFirmware, Action: "bmc.firmware.install", Summary: "bmc firmware install",
		Nodes: model.EventNodeList{{Node: "fw-1", State: model.EventNodeRunning, Stage: "pending", Ref: "JID_1"}},
	})

	pending := fakeJobs{"fw-1": {job("JID_1", schemas.PendingJobState, 0)}}
	h.checkFirmware(context.Background(), time.Now().UTC().Add(firmwareFollowFor+time.Minute), pending.read)

	got, err := h.DB.LoadEvent(event.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.State != model.EventStateDone || got.Nodes[0].State != model.EventNodeStalled {
		t.Errorf("event is %s with fw-1 %s, want it ended with fw-1 stalled", got.State, got.Nodes[0].State)
	}
	if got.Nodes[0].Message == "" {
		t.Error("the stalled node does not say why")
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
