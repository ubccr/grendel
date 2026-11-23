// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/facette/natsort"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/sqlstore/db"
	"github.com/ubccr/grendel/pkg/model"
)

// write runs fn in a transaction on the write connection, which is committed when fn returns nil and rolled back otherwise
func (s *SqlStore) write(fn func(tx *sql.Tx) error) error {
	tx, err := s.rw.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SqlStore) StoreEvent(event *model.Event) error {
	return s.write(func(tx *sql.Tx) error {
		if err := s.storeEvent(tx, event); err != nil {
			return err
		}
		return s.storeEventNodes(tx, event.ID, event.Nodes)
	})
}

func (s *SqlStore) storeEvent(tx *sql.Tx, event *model.Event) error {
	var ended *time.Time
	if event.EndedAt != nil {
		t := event.EndedAt.UTC()
		ended = &t
	}

	return s.q.EventUpsert(context.Background(), tx, db.EventUpsertParams{
		ID:        event.ID,
		Kind:      event.Kind,
		Action:    event.Action,
		Summary:   event.Summary,
		Subject:   event.Subject,
		Actor:     event.Actor,
		State:     event.State,
		Severity:  event.Severity,
		StartedAt: event.StartedAt.UTC(),
		UpdatedAt: event.UpdatedAt.UTC(),
		EndedAt:   ended,
		Detail:    event.Detail,
	})
}

func (s *SqlStore) storeEventNodes(tx *sql.Tx, id string, nodes model.EventNodeList) error {
	for _, n := range nodes {
		err := s.q.EventNodeUpsert(context.Background(), tx, db.EventNodeUpsertParams{
			EventID:   id,
			Node:      n.Node,
			State:     n.State,
			Stage:     n.Stage,
			Message:   n.Message,
			Ref:       n.Ref,
			UpdatedAt: n.UpdatedAt.UTC(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SqlStore) StoreEventNodes(id string, nodes model.EventNodeList, now time.Time) error {
	return s.write(func(tx *sql.Tx) error {
		if err := s.writeEventNodes(tx, id, nodes, now); err != nil {
			return err
		}
		return s.settleEvent(tx, id, now)
	})
}

// writeEventNodes writes rows of an event's nodes and says the event was updated, without settling it
func (s *SqlStore) writeEventNodes(tx *sql.Tx, id string, nodes model.EventNodeList, now time.Time) error {
	touched, err := s.q.EventTouch(context.Background(), tx, db.EventTouchParams{ID: id, UpdatedAt: now.UTC()})
	if err != nil {
		return err
	}
	if touched == 0 {
		return store.ErrNotFound
	}
	return s.storeEventNodes(tx, id, nodes)
}

// The counts are read and the event written in one transaction, so a node written meanwhile is never left out of the outcome
func (s *SqlStore) SettleEvent(id string, now time.Time) error {
	return s.write(func(tx *sql.Tx) error {
		return s.settleEvent(tx, id, now)
	})
}

func (s *SqlStore) settleEvent(tx *sql.Tx, id string, now time.Time) error {
	ctx := context.Background()
	row, err := s.q.EventFetch(ctx, tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return store.ErrNotFound
	}
	if err != nil {
		return err
	}
	event := eventFromRow(row)

	// a running event with a node still going stays as it is, which is found out without counting every node of a large event on each report
	if event.State == model.EventStateRunning {
		open, err := s.q.EventNodeOpen(ctx, tx, id)
		if err != nil || open {
			return err
		}
	}

	counts, err := s.q.EventNodeCounts(ctx, tx, []string{id})
	if err != nil {
		return err
	}
	for _, c := range counts {
		event.Counts.Add(c.State, int(c.Nodes))
	}

	if !event.Settle(now.UTC()) {
		return nil
	}
	return s.storeEvent(tx, event)
}

// trackEventNode writes how a tracked event is going on a node that reports in on its own, as a node provisioning does. The node joins the running event of open's kind, which open is stored as when there is none. A node that finished after the event it was in ended finishes in that event rather than joining a new one
func (s *SqlStore) trackEventNode(tx *sql.Tx, open *model.Event, node model.EventNode, now time.Time) error {
	ctx := context.Background()

	nodes := model.EventNodeList{node}

	if node.State.Finished() {
		latest, err := s.q.EventNodeLatest(ctx, tx, db.EventNodeLatestParams{Kind: open.Kind, Node: node.Node})
		switch {
		case err == nil:
			if err := s.writeEventNodes(tx, latest, nodes, now); err != nil {
				return err
			}
			return s.settleEvent(tx, latest, now)
		case !errors.Is(err, sql.ErrNoRows):
			return err
		}
	}

	id, err := s.q.EventTouchOpen(ctx, tx, db.EventTouchOpenParams{Kind: open.Kind, UpdatedAt: now.UTC()})
	if errors.Is(err, sql.ErrNoRows) {
		id = open.ID
		err = s.storeEvent(tx, open)
	}
	if err != nil {
		return err
	}
	if err := s.storeEventNodes(tx, id, nodes); err != nil {
		return err
	}

	// a node still going cannot end the running event it joins, so only a finished one is worth settling for
	if !node.State.Finished() {
		return nil
	}
	return s.settleEvent(tx, id, now)
}

func (s *SqlStore) LoadEvent(id string) (*model.Event, error) {
	row, err := s.q.EventFetch(context.Background(), s.ro, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	event := eventFromRow(row)

	nodes, err := s.q.EventNodes(context.Background(), s.ro, id)
	if err != nil {
		return nil, err
	}
	event.Nodes = model.EventNodeList{}
	for _, n := range nodes {
		addNode(event, n)
	}
	sortNodes(event)

	return event, nil
}

func (s *SqlStore) Events(filter model.EventFilter) (model.EventList, error) {
	ctx := context.Background()
	limit := filter.Limit
	if limit <= 0 {
		limit = model.DefaultEventLimit
	}
	limit = min(limit, model.MaxEventLimit)

	// the list query compares against the Before event's row, so with no such row it matches nothing and would read as the end of the list
	if filter.Before != "" {
		if _, err := s.q.EventFetch(ctx, s.ro, filter.Before); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("event %s: %w", filter.Before, store.ErrNotFound)
			}
			return nil, err
		}
	}

	rows, err := s.q.EventList(ctx, s.ro, db.EventListParams{
		State:  string(filter.State),
		Kind:   string(filter.Kind),
		Before: filter.Before,
		Max:    int64(limit),
	})
	if err != nil {
		return nil, err
	}

	events := make(model.EventList, 0, len(rows))
	byID := make(map[string]int, len(rows))
	ids := make([]string, 0, len(rows))
	for i, row := range rows {
		events = append(events, *eventFromRow(row))
		byID[row.ID] = i
		ids = append(ids, row.ID)
	}
	if len(ids) == 0 {
		return events, nil
	}

	if filter.Nodes {
		nodes, err := s.q.EventNodesIn(ctx, s.ro, ids)
		if err != nil {
			return nil, err
		}
		for _, n := range nodes {
			addNode(&events[byID[n.EventID]], n)
		}
		for i := range events {
			sortNodes(&events[i])
		}
		return events, nil
	}

	counts, err := s.q.EventNodeCounts(ctx, s.ro, ids)
	if err != nil {
		return nil, err
	}
	for _, c := range counts {
		countNode(&events[byID[c.EventID]], c.State, c.Stage, int(c.Nodes))
	}

	return events, nil
}

func addNode(event *model.Event, n db.EventNode) {
	event.Nodes = append(event.Nodes, model.EventNode{
		Node:      n.Node,
		State:     n.State,
		Stage:     n.Stage,
		Message:   n.Message,
		Ref:       n.Ref,
		UpdatedAt: n.UpdatedAt,
	})
	countNode(event, n.State, n.Stage, 1)
}

func sortNodes(event *model.Event) {
	slices.SortStableFunc(event.Nodes, func(a, b model.EventNode) int {
		switch {
		case natsort.Compare(a.Node, b.Node):
			return -1
		case natsort.Compare(b.Node, a.Node):
			return 1
		}
		return 0
	})
}

// countNode adds nodes at a state and stage to an event's counts, and the running ones to its stages
func countNode(event *model.Event, state model.EventNodeState, stage string, nodes int) {
	event.Counts.Add(state, nodes)
	if state == model.EventNodeRunning && stage != "" {
		if event.Stages == nil {
			event.Stages = map[string]int{}
		}
		event.Stages[stage] += nodes
	}
}

func (s *SqlStore) StallEventNodes(kind model.EventKind, before time.Time) (int64, error) {
	return s.q.EventNodesStall(context.Background(), s.rw, db.EventNodesStallParams{
		Kind:   kind,
		Before: before.UTC(),
	})
}

func (s *SqlStore) PruneEvents(before time.Time) (int64, error) {
	return s.q.EventPrune(context.Background(), s.rw, before.UTC())
}

func eventFromRow(row db.Event) *model.Event {
	return &model.Event{
		ID:        row.ID,
		Kind:      row.Kind,
		Action:    row.Action,
		Summary:   row.Summary,
		Subject:   row.Subject,
		Actor:     row.Actor,
		State:     row.State,
		Severity:  row.Severity,
		StartedAt: row.StartedAt,
		UpdatedAt: row.UpdatedAt,
		EndedAt:   row.EndedAt,
		Detail:    row.Detail,
	}
}
