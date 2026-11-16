// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import "time"

type Severity string

const (
	SeveritySuccess Severity = "success"
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

func (s Severity) String() string {
	return string(s)
}

// EventKind is what sort of thing an event records, which decides who keeps it up to date while it runs
type EventKind string

const (
	// EventKindAudit is a change someone made through the api, over by the time it is written
	EventKindAudit EventKind = "audit"
	// EventKindBmc is an action run against the BMCs of a set of nodes
	EventKindBmc EventKind = "bmc"
	// EventKindProvision is nodes network booting, kept up to date as they move through their provision
	EventKindProvision EventKind = "provision"
	// EventKindFirmware is firmware updates queued on BMCs, followed until their jobs finish
	EventKindFirmware EventKind = "firmware"
)

// EventState is whether an event is still being added to
type EventState string

const (
	EventStateRunning EventState = "running"
	EventStateDone    EventState = "done"
)

// EventNodeState is how an event is going on one node
type EventNodeState string

const (
	// EventNodePending is a node the event has not got to yet
	EventNodePending   EventNodeState = "pending"
	EventNodeRunning   EventNodeState = "running"
	EventNodeSucceeded EventNodeState = "succeeded"
	EventNodeFailed    EventNodeState = "failed"
	// EventNodeStalled is a node that stopped reporting before it finished
	EventNodeStalled EventNodeState = "stalled"
)

// Finished reports whether the event is over for the node, however it went
func (s EventNodeState) Finished() bool {
	return s != EventNodePending && s != EventNodeRunning
}

// EventActorSystem is the actor of an event nobody asked for, such as a node network booting
const EventActorSystem = "system"

// Event is something that was done to grendel or its nodes. Most are done when they are written. A tracked event is written while it is running and updated as its nodes report in, then closed
//
// Event is only ever a response, so validate:"required" marks the fields always sent as required in the OpenAPI schema, which keeps them from being optional in the generated clients, and never validates anything
type Event struct {
	// ID is unique and sorts by the second the event started. Events are ordered by StartedAt, with the ID breaking ties
	ID   string    `json:"id" validate:"required"`
	Kind EventKind `json:"kind" validate:"required" enum:"audit,bmc,provision,firmware"`
	// Action names what was done for a program to read, such as node.save or bmc.power.cycle
	Action string `json:"action" validate:"required"`
	// Summary says what was done for a person to read, without how it went, which the counts say
	Summary string `json:"summary" validate:"required"`
	// Subject is what the event was about when it has no rows of nodes: a nodeset, or the names of users or images
	Subject string `json:"subject,omitempty"`
	// Actor is the user who did it, or EventActorSystem
	Actor     string     `json:"actor" validate:"required"`
	State     EventState `json:"state" validate:"required" enum:"running,done"`
	Severity  Severity   `json:"severity" validate:"required" enum:"success,info,warning,error"`
	StartedAt time.Time  `json:"started_at" validate:"required"`
	UpdatedAt time.Time  `json:"updated_at" validate:"required"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	// Detail is what only this kind of event has, such as a bmc action's run id
	Detail map[string]string `json:"detail,omitempty"`
	// Counts are the event's nodes by how it is going on them. An event with no rows of nodes has none
	Counts EventCounts `json:"counts" validate:"required"`
	// Stages are the nodes still running by the stage they are at, such as booting for a provision
	Stages map[string]int `json:"stages,omitempty"`
	// Nodes are how the event went on each node, in natural order. Only a single event has them
	Nodes EventNodeList `json:"nodes,omitempty"`
}

type EventList []Event

// EventCounts are an event's nodes by state
type EventCounts struct {
	Total     int `json:"total" validate:"required"`
	Pending   int `json:"pending" validate:"required"`
	Running   int `json:"running" validate:"required"`
	Succeeded int `json:"succeeded" validate:"required"`
	Failed    int `json:"failed" validate:"required"`
	Stalled   int `json:"stalled" validate:"required"`
}

// Open is how many nodes the event is not over for
func (c EventCounts) Open() int {
	return c.Pending + c.Running
}

// Severity is how an event with these counts went once it is over: an error when nodes failed and none succeeded, a warning when some failed or stalled, and a success otherwise
func (c EventCounts) Severity() Severity {
	switch {
	case c.Failed > 0 && c.Succeeded == 0:
		return SeverityError
	case c.Failed > 0 || c.Stalled > 0:
		return SeverityWarning
	}
	return SeveritySuccess
}

// Add counts n nodes at a state
func (c *EventCounts) Add(state EventNodeState, n int) {
	c.Total += n
	switch state {
	case EventNodePending:
		c.Pending += n
	case EventNodeRunning:
		c.Running += n
	case EventNodeSucceeded:
		c.Succeeded += n
	case EventNodeFailed:
		c.Failed += n
	case EventNodeStalled:
		c.Stalled += n
	}
}

// Settle brings an event up to date with its Counts: a running event ends once it is over for every node, and one that has ended takes the severity a node reporting late gives it. An event with no nodes is left as it is. It reports whether anything changed
func (e *Event) Settle(now time.Time) bool {
	if e.Counts.Total == 0 {
		return false
	}
	running := e.State == EventStateRunning
	if running && e.Counts.Open() > 0 {
		return false
	}
	severity := e.Counts.Severity()
	if !running && severity == e.Severity {
		return false
	}

	e.Severity = severity
	e.UpdatedAt = now
	if running {
		e.State = EventStateDone
		e.EndedAt = &now
	}
	return true
}

// EventNode is how an event went on one node
type EventNode struct {
	Node  string         `json:"node" validate:"required"`
	State EventNodeState `json:"state" validate:"required" enum:"pending,running,succeeded,failed,stalled"`
	// Stage is the step a running node is at, or the last it reached
	Stage   string `json:"stage,omitempty"`
	Message string `json:"message,omitempty"`
	// Ref is what the event follows the node by, such as the id of the job a firmware update queued on its BMC
	Ref       string    `json:"ref,omitempty"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

type EventNodeList []EventNode

// EventFilter picks the events to list. The zero value is the latest events of every kind
type EventFilter struct {
	State EventState
	Kind  EventKind
	// Before keeps the events older than the one with this id, for reading further back
	Before string
	// Limit is the most events returned, DefaultEventLimit when zero
	Limit int
	// Nodes fills in each event's nodes, which a list otherwise leaves out. It is for grendel's own use, as the api lists events without them
	Nodes bool
}

const (
	DefaultEventLimit = 100
	MaxEventLimit     = 1000
)
