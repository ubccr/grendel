// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/facette/natsort"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/sqlstore/db"
	"github.com/ubccr/grendel/pkg/model"
)

// latestChunk is how many nodes are looked up in one query, which keeps it well under SQLite's limit on parameters however many nodes report at once
const latestChunk = 500

// StoreEvent writes an event, replacing what can change about one already stored with the same ID, and writes the rows of its Nodes, leaving its other nodes as they were
func (s *SqlStore) StoreEvent(event *model.Event) error {
	tx, err := s.rw.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.storeEvent(tx, event); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SqlStore) storeEvent(tx *sql.Tx, event *model.Event) error {
	detail := "{}"
	if len(event.Detail) > 0 {
		raw, err := json.Marshal(event.Detail)
		if err != nil {
			return err
		}
		detail = string(raw)
	}

	counts := "{}"
	if event.Counts.Total > 0 {
		raw, err := json.Marshal(event.Counts)
		if err != nil {
			return err
		}
		counts = string(raw)
	}

	ended := sql.NullTime{}
	if event.EndedAt != nil {
		ended = sql.NullTime{Time: event.EndedAt.UTC(), Valid: true}
	}

	err := s.q.EventUpsert(context.Background(), tx, db.EventUpsertParams{
		ID:        event.ID,
		Kind:      string(event.Kind),
		Action:    event.Action,
		Summary:   event.Summary,
		Subject:   event.Subject,
		Actor:     event.Actor,
		State:     string(event.State),
		Severity:  string(event.Severity),
		StartedAt: event.StartedAt.UTC(),
		UpdatedAt: event.UpdatedAt.UTC(),
		EndedAt:   ended,
		Detail:    detail,
		Counts:    counts,
	})
	if err != nil {
		return err
	}

	return s.storeEventNodes(tx, event.ID, event.Nodes)
}

func (s *SqlStore) storeEventNodes(tx *sql.Tx, id string, nodes model.EventNodeList) error {
	for _, n := range nodes {
		err := s.q.EventNodeUpsert(context.Background(), tx, db.EventNodeUpsertParams{
			EventID:   id,
			Node:      n.Node,
			State:     string(n.State),
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

// StoreEventNodes writes how an event is going on the given nodes, leaving its other nodes and the event itself as they were, but for when it was last updated. It is store.ErrNotFound when there is no such event
func (s *SqlStore) StoreEventNodes(id string, nodes model.EventNodeList, now time.Time) error {
	tx, err := s.rw.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.touchEvent(tx, id, now); err != nil {
		return err
	}
	if err := s.storeEventNodes(tx, id, nodes); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SqlStore) touchEvent(tx *sql.Tx, id string, now time.Time) error {
	touched, err := s.q.EventTouch(context.Background(), tx, db.EventTouchParams{ID: id, UpdatedAt: now.UTC()})
	if err != nil {
		return err
	}
	if touched == 0 {
		return store.ErrNotFound
	}
	return nil
}

// SettleEvent brings an event up to date with how it is going on its nodes, as model.Event.Settle does. The counts are read and the event written in one transaction, so a node written meanwhile is never left out of the outcome. It is store.ErrNotFound when there is no such event
func (s *SqlStore) SettleEvent(id string, now time.Time) error {
	tx, err := s.rw.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.settleEvent(tx, id, now); err != nil {
		return err
	}
	return tx.Commit()
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
	event, err := eventFromRow(row)
	if err != nil {
		return err
	}

	counts, err := s.q.EventNodeCounts(ctx, tx, []string{id})
	if err != nil {
		return err
	}
	stored := event.Counts
	event.Counts = model.EventCounts{}
	for _, c := range counts {
		event.Counts.Add(model.EventNodeState(c.State), int(c.Nodes))
	}

	// the counts are stored even when nothing else changed, since a finished event is listed with the counts it was stored with
	if !event.Settle(now.UTC()) && event.Counts == stored {
		return nil
	}
	return s.storeEvent(tx, event)
}

// TrackEventNodes writes how a tracked event is going on nodes that report in on their own, as nodes provisioning do, all in one transaction. A node joins the running event of open's kind, which open is stored as when there is none. A node that finished after the event it was in ended finishes in that event rather than joining a new one. Every event written to is then settled
func (s *SqlStore) TrackEventNodes(open *model.Event, nodes model.EventNodeList, now time.Time) error {
	if len(nodes) == 0 {
		return nil
	}
	ctx := context.Background()
	now = now.UTC()

	tx, err := s.rw.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var running string
	row, err := s.q.EventOpen(ctx, tx, string(open.Kind))
	switch {
	case err == nil:
		running = row.ID
	case !errors.Is(err, sql.ErrNoRows):
		return err
	}

	// the event each finished node was last in, which is where a node that went quiet and then finished belongs
	var finished []string
	for _, n := range nodes {
		if n.State.Finished() {
			finished = append(finished, n.Node)
		}
	}
	latest := make(map[string]string, len(finished))
	for chunk := range slices.Chunk(finished, latestChunk) {
		rows, err := s.q.EventNodesLatest(ctx, tx, db.EventNodesLatestParams{Kind: string(open.Kind), Nodes: chunk})
		if err != nil {
			return err
		}
		for _, r := range rows {
			if _, ok := latest[r.Node]; !ok {
				latest[r.Node] = r.ID
			}
		}
	}

	byEvent := map[string]model.EventNodeList{}
	for _, n := range nodes {
		id, ok := latest[n.Node]
		if !ok {
			if running == "" {
				if err := s.storeEvent(tx, open); err != nil {
					return err
				}
				running = open.ID
			}
			id = running
		}
		byEvent[id] = append(byEvent[id], n)
	}

	for id, rows := range byEvent {
		if err := s.touchEvent(tx, id, now); err != nil {
			return err
		}
		if err := s.storeEventNodes(tx, id, rows); err != nil {
			return err
		}
		if err := s.settleEvent(tx, id, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// LoadEvent returns the event with the given ID with its nodes, or store.ErrNotFound
func (s *SqlStore) LoadEvent(id string) (*model.Event, error) {
	row, err := s.q.EventFetch(context.Background(), s.ro, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	event, err := eventFromRow(row)
	if err != nil {
		return nil, err
	}

	nodes, err := s.q.EventNodes(context.Background(), s.ro, id)
	if err != nil {
		return nil, err
	}
	// with every node at hand the counts are made from them rather than taken as stored
	event.Counts = model.EventCounts{}
	event.Nodes = model.EventNodeList{}
	for _, n := range nodes {
		addNode(event, n)
	}
	sortNodes(event)

	return event, nil
}

// Events returns the events the filter picks, newest first, with their counts, and with their nodes when the filter asks. A finished event's counts are the ones stored when it was last settled, and only running events have their nodes counted, so listing costs the same however many nodes old events had
func (s *SqlStore) Events(filter model.EventFilter) (model.EventList, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = model.DefaultEventLimit
	}
	limit = min(limit, model.MaxEventLimit)

	rows, err := s.q.EventList(context.Background(), s.ro, db.EventListParams{
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
	var running []string
	for i, row := range rows {
		event, err := eventFromRow(row)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
		byID[row.ID] = i
		ids = append(ids, row.ID)
		if event.State == model.EventStateRunning {
			running = append(running, row.ID)
		}
	}
	if len(ids) == 0 {
		return events, nil
	}

	if filter.Nodes {
		for i := range events {
			events[i].Counts = model.EventCounts{}
		}
		nodes, err := s.q.EventNodesIn(context.Background(), s.ro, ids)
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

	if len(running) == 0 {
		return events, nil
	}
	counts, err := s.q.EventNodeCounts(context.Background(), s.ro, running)
	if err != nil {
		return nil, err
	}
	for _, c := range counts {
		countNode(&events[byID[c.EventID]], model.EventNodeState(c.State), c.Stage, int(c.Nodes))
	}

	return events, nil
}

// addNode adds a node's row to an event and to its counts
func addNode(event *model.Event, n db.EventNode) {
	state := model.EventNodeState(n.State)
	event.Nodes = append(event.Nodes, model.EventNode{
		Node:      n.Node,
		State:     state,
		Stage:     n.Stage,
		Message:   n.Message,
		Ref:       n.Ref,
		UpdatedAt: n.UpdatedAt,
	})
	countNode(event, state, n.Stage, 1)
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

// StallEventNodes marks the nodes of running events of a kind as stalled when they were last updated before the given time, and returns how many
func (s *SqlStore) StallEventNodes(kind model.EventKind, before time.Time) (int64, error) {
	return s.q.EventNodesStall(context.Background(), s.rw, db.EventNodesStallParams{
		Kind:   string(kind),
		Before: before.UTC(),
	})
}

// PruneEvents deletes the finished events that started before the given time, and returns how many
func (s *SqlStore) PruneEvents(before time.Time) (int64, error) {
	return s.q.EventPrune(context.Background(), s.rw, before.UTC())
}

func eventFromRow(row db.Event) (*model.Event, error) {
	event := &model.Event{
		ID:        row.ID,
		Kind:      model.EventKind(row.Kind),
		Action:    row.Action,
		Summary:   row.Summary,
		Subject:   row.Subject,
		Actor:     row.Actor,
		State:     model.EventState(row.State),
		Severity:  model.Severity(row.Severity),
		StartedAt: row.StartedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.EndedAt.Valid {
		event.EndedAt = &row.EndedAt.Time
	}
	if row.Detail != "" {
		if err := json.Unmarshal([]byte(row.Detail), &event.Detail); err != nil {
			return nil, err
		}
	}
	// a running event is counted from its nodes, so what it was stored with is left out
	if row.Counts != "" && event.State != model.EventStateRunning {
		if err := json.Unmarshal([]byte(row.Counts), &event.Counts); err != nil {
			return nil, err
		}
	}
	return event, nil
}
