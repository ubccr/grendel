// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package storetest

import (
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
)

func testEvent(id string, kind model.EventKind, state model.EventState, started time.Time) *model.Event {
	return &model.Event{
		ID:        id,
		Kind:      kind,
		Action:    "test",
		Summary:   "event " + id,
		Actor:     "admin",
		State:     state,
		Severity:  model.SeverityInfo,
		StartedAt: started,
		UpdatedAt: started,
	}
}

func (s *StoreTestSuite) TestEvents() {
	defer s.db.Close()
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)

	_, err := s.db.LoadEvent("missing")
	s.Assert().ErrorIs(err, store.ErrNotFound)

	list, err := s.db.Events(model.EventFilter{})
	s.Require().NoError(err)
	s.Assert().NotNil(list, "no events is an empty list, so it is sent as [] rather than null")
	s.Assert().Empty(list)

	audit := testEvent("a1", model.EventKindAudit, model.EventStateDone, now)
	audit.Subject = "cpn-[01-02]"
	audit.Detail = model.EventDetail{"run": "r1"}
	s.Require().NoError(s.db.StoreEvent(audit))

	run := testEvent("b2", model.EventKindBmc, model.EventStateRunning, now.Add(time.Minute))
	run.Nodes = model.EventNodeList{
		{Node: "cpn-10", State: model.EventNodePending, UpdatedAt: now},
		{Node: "cpn-2", State: model.EventNodeRunning, Stage: "booting", UpdatedAt: now},
		{Node: "cpn-3", State: model.EventNodeRunning, Stage: "booting", UpdatedAt: now},
	}
	s.Require().NoError(s.db.StoreEvent(run))

	// writing some of an event's nodes leaves the others as they were
	s.Require().NoError(s.db.StoreEventNodes("b2", model.EventNodeList{{Node: "cpn-2", State: model.EventNodeFailed, Message: "timed out", UpdatedAt: now.Add(time.Minute)}}, now.Add(time.Minute)))
	s.Assert().ErrorIs(s.db.StoreEventNodes("missing", nil, now), store.ErrNotFound)

	loaded, err := s.db.LoadEvent("b2")
	s.Require().NoError(err)
	s.Assert().Equal(model.EventCounts{Total: 3, Pending: 1, Running: 1, Failed: 1}, loaded.Counts)
	s.Assert().Equal(map[string]int{"booting": 1}, loaded.Stages)
	s.Require().Len(loaded.Nodes, 3)
	s.Assert().Equal([]string{"cpn-2", "cpn-3", "cpn-10"}, []string{loaded.Nodes[0].Node, loaded.Nodes[1].Node, loaded.Nodes[2].Node}, "nodes are in natural order")
	s.Assert().Equal("timed out", loaded.Nodes[0].Message)
	s.Assert().True(loaded.UpdatedAt.Equal(now.Add(time.Minute)), "writing nodes says the event was updated")

	list, err = s.db.Events(model.EventFilter{})
	s.Require().NoError(err)
	s.Require().Len(list, 2)
	s.Assert().Equal("b2", list[0].ID, "newest first")
	s.Assert().Equal(loaded.Counts, list[0].Counts)
	s.Assert().Empty(list[0].Nodes, "a list leaves the nodes out")
	withNodes, err := s.db.Events(model.EventFilter{Nodes: true})
	s.Require().NoError(err)
	s.Assert().Equal(loaded.Nodes, withNodes[0].Nodes, "unless asked")
	s.Assert().Equal(loaded.Counts, withNodes[0].Counts)
	s.Assert().Equal("cpn-[01-02]", list[1].Subject)
	s.Assert().Equal(model.EventDetail{"run": "r1"}, list[1].Detail)
	s.Assert().Zero(list[1].Counts)

	for name, c := range map[string]struct {
		filter model.EventFilter
		want   []string
	}{
		"state":  {model.EventFilter{State: model.EventStateRunning}, []string{"b2"}},
		"kind":   {model.EventFilter{Kind: model.EventKindAudit}, []string{"a1"}},
		"before": {model.EventFilter{Before: "b2"}, []string{"a1"}},
		"limit":  {model.EventFilter{Limit: 1}, []string{"b2"}},
	} {
		got, err := s.db.Events(c.filter)
		s.Require().NoError(err, name)
		ids := []string{}
		for _, e := range got {
			ids = append(ids, e.ID)
		}
		s.Assert().Equal(c.want, ids, name)
	}

	// only nodes of running events of the kind that went quiet are stalled
	stalled, err := s.db.StallEventNodes(model.EventKindProvision, now.Add(time.Hour))
	s.Require().NoError(err)
	s.Assert().Zero(stalled)
	stalled, err = s.db.StallEventNodes(model.EventKindBmc, now.Add(time.Second))
	s.Require().NoError(err)
	s.Assert().EqualValues(2, stalled, "the pending and the running node, and not the one updated since")
	loaded, err = s.db.LoadEvent("b2")
	s.Require().NoError(err)
	s.Assert().Equal(model.EventCounts{Total: 3, Failed: 1, Stalled: 2}, loaded.Counts)

	// closing replaces what can change and keeps the rest
	ended := now.Add(2 * time.Minute)
	run.Nodes = nil
	run.State = model.EventStateDone
	run.Severity = model.SeverityError
	run.EndedAt = &ended
	run.Kind = model.EventKindAudit
	s.Require().NoError(s.db.StoreEvent(run))
	loaded, err = s.db.LoadEvent("b2")
	s.Require().NoError(err)
	s.Assert().Equal(model.EventStateDone, loaded.State)
	s.Assert().Equal(model.SeverityError, loaded.Severity)
	s.Require().NotNil(loaded.EndedAt)
	s.Assert().True(loaded.EndedAt.Equal(ended))
	s.Assert().Equal(model.EventKindBmc, loaded.Kind, "an event's kind is set once")
	s.Assert().Len(loaded.Nodes, 3)

	// pruning takes finished events and their nodes, and leaves running ones however old
	old := testEvent("00", model.EventKindProvision, model.EventStateRunning, now.Add(-48*time.Hour))
	s.Require().NoError(s.db.StoreEvent(old))
	pruned, err := s.db.PruneEvents(now.Add(30 * time.Second))
	s.Require().NoError(err)
	s.Assert().EqualValues(1, pruned)
	list, err = s.db.Events(model.EventFilter{})
	s.Require().NoError(err)
	s.Require().Len(list, 2)
	s.Assert().Equal([]string{"b2", "00"}, []string{list[0].ID, list[1].ID})

	// reading back from an event that is not there is an error, not the end of the list
	_, err = s.db.Events(model.EventFilter{Before: "a1"})
	s.Assert().ErrorIs(err, store.ErrNotFound)
}

func (s *StoreTestSuite) TestUpdateProvisionStatus() {
	defer s.db.Close()
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)

	hosts := make([]*model.Host, 8)
	for i := range hosts {
		hosts[i] = tests.HostFactory.MustCreate().(*model.Host)
		hosts[i].Name = fmt.Sprintf("n%d", i)
		s.Require().NoError(s.db.StoreHost(hosts[i]))
	}
	// report moves a host to a state and writes its row of the provisioning, which is opened under the given id when none is running
	report := func(id string, host *model.Host, state model.ProvisionState, at time.Time) error {
		open := testEvent(id, model.EventKindProvision, model.EventStateRunning, at)
		return s.db.UpdateProvisionStatus(host.UID.String(), open, func(status *model.ProvisionStatus) (*model.EventNode, error) {
			status.State = state
			status.StartedAt = now
			status.UpdatedAt = at
			row := &model.EventNode{Node: host.Name, State: model.EventNodeRunning, Stage: string(state), UpdatedAt: at}
			switch state {
			case model.ProvisionStateCompleted:
				row.State = model.EventNodeSucceeded
			case model.ProvisionStateFailed:
				row.State = model.EventNodeFailed
			}
			return row, nil
		})
	}
	provisions := func() model.EventList {
		list, err := s.db.Events(model.EventFilter{Kind: model.EventKindProvision})
		s.Require().NoError(err)
		return list
	}

	_, err := s.db.LoadProvisionStatus(hosts[0].UID.String())
	s.Assert().ErrorIs(err, store.ErrNotFound)

	// nodes reporting at once all land in the one event
	errs := make(chan error, len(hosts))
	for i, host := range hosts {
		go func() { errs <- report(fmt.Sprintf("p%d", i), host, model.ProvisionStatePXE, now) }()
	}
	for range hosts {
		s.Require().NoError(<-errs)
	}
	list := provisions()
	s.Require().Len(list, 1)
	first := list[0].ID
	s.Assert().Equal(model.EventCounts{Total: 8, Running: 8}, list[0].Counts)
	s.Assert().Equal(map[string]int{"pxe": 8}, list[0].Stages)
	status, err := s.db.LoadProvisionStatus(hosts[0].UID.String())
	s.Require().NoError(err)
	s.Assert().Equal(model.ProvisionStatePXE, status.State)

	// update is given what is stored, and an error from it leaves that and the event as they were
	refused := errors.New("refused")
	err = s.db.UpdateProvisionStatus(hosts[0].UID.String(), testEvent("p8", model.EventKindProvision, model.EventStateRunning, now), func(status *model.ProvisionStatus) (*model.EventNode, error) {
		s.Assert().Equal(model.ProvisionStatePXE, status.State)
		status.State = model.ProvisionStateFailed
		return &model.EventNode{Node: hosts[0].Name, State: model.EventNodeFailed, UpdatedAt: now}, refused
	})
	s.Assert().ErrorIs(err, refused)
	status, err = s.db.LoadProvisionStatus(hosts[0].UID.String())
	s.Require().NoError(err)
	s.Assert().Equal(model.ProvisionStatePXE, status.State)
	s.Assert().Equal(model.EventCounts{Total: 8, Running: 8}, provisions()[0].Counts)

	// a status changed with no node to track leaves the events alone
	err = s.db.UpdateProvisionStatus(hosts[0].UID.String(), testEvent("p8", model.EventKindProvision, model.EventStateRunning, now), func(status *model.ProvisionStatus) (*model.EventNode, error) {
		status.Message = "noted"
		return nil, nil
	})
	s.Require().NoError(err)
	status, err = s.db.LoadProvisionStatus(hosts[0].UID.String())
	s.Require().NoError(err)
	s.Assert().Equal("noted", status.Message)
	s.Assert().Equal(model.EventCounts{Total: 8, Running: 8}, provisions()[0].Counts)

	// a running event of another kind is not joined
	s.Require().NoError(s.db.StoreEvent(testEvent("f1", model.EventKindFirmware, model.EventStateRunning, now)))
	s.Require().NoError(report("p9", hosts[0], model.ProvisionStateCompleted, now))
	s.Assert().Len(provisions(), 1)

	// the event ends in the same write as its last node finishing
	ended := now.Add(time.Minute)
	for _, host := range hosts[1:7] {
		s.Require().NoError(report("p9", host, model.ProvisionStateCompleted, ended))
	}
	s.Assert().Equal(model.EventStateRunning, provisions()[0].State)
	s.Require().NoError(report("p9", hosts[7], model.ProvisionStateFailed, ended))
	loaded, err := s.db.LoadEvent(first)
	s.Require().NoError(err)
	s.Assert().Equal(model.EventStateDone, loaded.State)
	s.Assert().Equal(model.SeverityWarning, loaded.Severity)
	s.Require().NotNil(loaded.EndedAt)
	s.Assert().True(loaded.EndedAt.Equal(ended))
	s.Assert().Equal(model.EventCounts{Total: 8, Succeeded: 7, Failed: 1}, provisions()[0].Counts)

	// after that a node starting opens a new event, while one finishing late finishes in the event it was in
	later := now.Add(2 * time.Minute)
	s.Require().NoError(report("p10", hosts[1], model.ProvisionStatePXE, later))
	s.Require().NoError(report("p11", hosts[7], model.ProvisionStateCompleted, later))
	list = provisions()
	s.Require().Len(list, 2)
	s.Assert().Equal("p10", list[0].ID)
	s.Assert().Equal(model.EventCounts{Total: 1, Running: 1}, list[0].Counts)
	s.Assert().Equal(first, list[1].ID)
	s.Assert().Equal(model.SeveritySuccess, list[1].Severity)
	s.Assert().Equal(model.EventCounts{Total: 8, Succeeded: 8}, list[1].Counts, "and the late node is in them")

	s.Assert().ErrorIs(report("p12", &model.Host{Name: "gone", UID: ksuid.New()}, model.ProvisionStatePXE, later), store.ErrNotFound)
	s.Assert().ErrorIs(s.db.SettleEvent("missing", now), store.ErrNotFound)
}
