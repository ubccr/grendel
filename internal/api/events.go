// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/pkg/model"
)

// GetEvents lists the latest events, newest first, with how many of their nodes are at each state. The nodes themselves are in GetEvent
func (h *Handler) GetEvents(c fuego.ContextNoBody) (model.EventList, error) {
	limit, err := c.QueryParamIntErr("limit")
	if err != nil {
		return nil, badRequest(err, "limit must be a number")
	}

	events, err := h.DB.Events(model.EventFilter{
		State:  model.EventState(c.QueryParam("state")),
		Kind:   model.EventKind(c.QueryParam("kind")),
		Before: c.QueryParam("before"),
		Limit:  limit,
	})
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to load events"}
	}
	return events, nil
}

// GetEvent returns one event with how it went on each of its nodes
func (h *Handler) GetEvent(c fuego.ContextNoBody) (*model.Event, error) {
	event, err := h.DB.LoadEvent(c.PathParam("id"))
	if errors.Is(err, store.ErrNotFound) {
		return nil, fuego.HTTPError{Err: err, Status: http.StatusNotFound, Title: "Not Found", Detail: "no event with that id"}
	}
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to load the event"}
	}
	return event, nil
}

// actor is the user a request was made as. Every authenticated route has one; a request without is recorded as the system's
func actor(ctx context.Context) string {
	if username, ok := ctx.Value(ContextKeyUsername).(string); ok && username != "" {
		return username
	}
	return model.EventActorSystem
}

// audit records a change someone made through the api. action names it for programs, summary says it for people, and subject is what it was made to
func (h *Handler) audit(ctx context.Context, severity model.Severity, action, summary, subject string) {
	h.Events.Record(model.Event{
		Kind:     model.EventKindAudit,
		Action:   action,
		Summary:  summary,
		Subject:  subject,
		Actor:    actor(ctx),
		Severity: severity,
	})
}
