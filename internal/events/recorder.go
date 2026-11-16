// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package events records what is done to grendel and its nodes. Every service writes its events through a Recorder, so the api, the provision server and DHCP all add to the one log in the database
package events

import (
	"context"
	"errors"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/pkg/model"
)

var log = logger.GetLogger("EVENTS")

// how long a node of a running event can go without an update before it counts as stalled. A node with no beacon never says its provision finished, and a bmc action whose server died never reports at all
const (
	ProvisionStalledAfter = 10 * time.Minute
	bmcStalledAfter       = time.Hour
)

const (
	// FlushInterval is how often the reports of nodes provisioning are written. They are held in memory until then, so a whole cluster booting at once costs one write a second rather than one per report
	FlushInterval = time.Second
	tidyInterval  = time.Minute
)

func init() {
	viper.SetDefault("events.retention", "2160h")
	viper.SetDefault("bmc.firmware_timeout", "2h")
}

// Recorder writes events to the store. Recording never gets in the way of what is being recorded, so a failure is only logged
type Recorder struct {
	DB store.Store
}

func New(db store.Store) *Recorder {
	return &Recorder{DB: db}
}

// start fills in what every new event has
func start(event *model.Event, now time.Time) {
	id, err := ksuid.NewRandomWithTime(now)
	if err != nil {
		id = ksuid.New()
	}
	event.ID = id.String()
	event.StartedAt = now
	event.UpdatedAt = now
	if event.Actor == "" {
		event.Actor = model.EventActorSystem
	}
}

// Record writes an event that is already over, with the severity given
func (r *Recorder) Record(event model.Event) *model.Event {
	now := time.Now().UTC()
	start(&event, now)
	event.State = model.EventStateDone
	event.EndedAt = &now
	r.store(&event)
	return &event
}

// Open writes an event that is still running, with its Nodes, to be updated with SetNodes and ended by Settle once it is over for them all
func (r *Recorder) Open(event model.Event) *model.Event {
	now := time.Now().UTC()
	start(&event, now)
	event.State = model.EventStateRunning
	event.Severity = model.SeverityInfo
	for i := range event.Nodes {
		if event.Nodes[i].UpdatedAt.IsZero() {
			event.Nodes[i].UpdatedAt = now
		}
	}
	r.store(&event)
	event.Nodes = nil
	return &event
}

// SetNodes writes how the event is going on the given nodes, leaving its other nodes as they were. It works on an event that has ended too, for a node that reports late
func (r *Recorder) SetNodes(event *model.Event, nodes ...model.EventNode) {
	now := time.Now().UTC()
	for i := range nodes {
		if nodes[i].UpdatedAt.IsZero() {
			nodes[i].UpdatedAt = now
		}
	}
	if err := r.DB.StoreEventNodes(event.ID, nodes, now); err != nil {
		log.Warnf("failed to store the nodes of event %q: %s", event.Summary, err)
	}
}

// Settle ends a running event once it is over for every one of its nodes, and brings the severity of one already ended up to date with a node that reported late
func (r *Recorder) Settle(event *model.Event) {
	if err := r.DB.SettleEvent(event.ID, time.Now().UTC()); err != nil {
		log.Warnf("failed to settle event %q: %s", event.Summary, err)
	}
}

func (r *Recorder) store(event *model.Event) {
	if err := r.DB.StoreEvent(event); err != nil {
		log.Warnf("failed to store event %q: %s", event.Summary, err)
	}
}

// provisionQueue holds the latest report of each node provisioning until the next flush writes them all at once
type provisionQueue struct {
	mu    sync.Mutex
	nodes map[string]model.EventNode
}

// queues are the provision queues by store, so each store's reports are written to it and tests with a store each stay apart
var queues sync.Map

func queueFor(db store.Store) *provisionQueue {
	q, _ := queues.LoadOrStore(db, &provisionQueue{nodes: map[string]model.EventNode{}})
	return q.(*provisionQueue)
}

// TrackProvision records how a node's provision is going, which joins it to the provisioning under way. It is called on a node's boot path, so it only notes the report in memory; Flush writes it with every other report since the last, within FlushInterval while Run is running. A later report from the same node replaces one not yet written
func TrackProvision(db store.Store, node model.EventNode) {
	if node.UpdatedAt.IsZero() {
		node.UpdatedAt = time.Now().UTC()
	}
	q := queueFor(db)
	q.mu.Lock()
	q.nodes[node.Node] = node
	q.mu.Unlock()
}

// Flush writes the provision reports noted since the last flush, in one transaction. Reports it fails to write are kept for the next flush, unless a newer report from the same node has come in meanwhile
func (r *Recorder) Flush() {
	q := queueFor(r.DB)
	q.mu.Lock()
	if len(q.nodes) == 0 {
		q.mu.Unlock()
		return
	}
	nodes := q.nodes
	q.nodes = make(map[string]model.EventNode, len(nodes))
	q.mu.Unlock()

	now := time.Now().UTC()
	open := model.Event{Kind: model.EventKindProvision, Action: "provision", Summary: "Provisioning", State: model.EventStateRunning, Severity: model.SeverityInfo}
	start(&open, now)
	if err := r.DB.TrackEventNodes(&open, slices.Collect(maps.Values(nodes)), now); err != nil {
		log.Warnf("failed to write the reports of %d provisioning nodes, retrying: %s", len(nodes), err)
		q.mu.Lock()
		for name, n := range nodes {
			if _, newer := q.nodes[name]; !newer {
				q.nodes[name] = n
			}
		}
		q.mu.Unlock()
	}
}

// Tidy marks the nodes of running events that went quiet as stalled, ends the events that leaves with nothing running, and deletes the events older than events.retention. It only changes what is out of date, so more than one service can run it
func (r *Recorder) Tidy(now time.Time) {
	stalledAfter := map[model.EventKind]time.Duration{
		model.EventKindProvision: ProvisionStalledAfter,
		model.EventKindBmc:       bmcStalledAfter,
		model.EventKindFirmware:  viper.GetDuration("bmc.firmware_timeout"),
	}
	for kind, after := range stalledAfter {
		if after <= 0 {
			continue
		}
		if _, err := r.DB.StallEventNodes(kind, now.Add(-after)); err != nil {
			log.Warnf("failed to mark stalled %s nodes: %s", kind, err)
		}
	}

	running, err := r.DB.Events(model.EventFilter{State: model.EventStateRunning, Limit: model.MaxEventLimit})
	if err != nil {
		log.Warnf("failed to list running events: %s", err)
	}
	for _, event := range running {
		if event.Counts.Total > 0 && event.Counts.Open() == 0 {
			r.Settle(&event)
		}
	}

	if keep := viper.GetDuration("events.retention"); keep > 0 {
		pruned, err := r.DB.PruneEvents(now.Add(-keep))
		if err != nil {
			log.Warnf("failed to prune events: %s", err)
		} else if pruned > 0 {
			log.Infof("Deleted %d events older than %s", pruned, keep)
		}
	}
}

// Run writes the provision reports every FlushInterval and tidies the events every minute, until the context ends, when it writes what reports are left
func (r *Recorder) Run(ctx context.Context) error {
	flush := time.NewTicker(FlushInterval)
	defer flush.Stop()
	tidy := time.NewTicker(tidyInterval)
	defer tidy.Stop()

	r.Tidy(time.Now().UTC())
	for {
		select {
		case <-ctx.Done():
			r.Flush()
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
			return ctx.Err()
		case <-flush.C:
			r.Flush()
		case <-tidy.C:
			r.Tidy(time.Now().UTC())
		}
	}
}
