// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/pkg/model"
)

func (h *Handler) GetEvents(c fuego.ContextNoBody) (model.EventList, error) {
	return h.Events.GetEvents(), nil
}

func (h *Handler) writeEvent(ctx context.Context, severity, msg string, jobMessages ...model.JobMessage) {
	username, ok := ctx.Value(ContextKeyUsername).(string)
	if !ok {
		log.Warn("failed to get username from http context, ignoring writeEvent")
		return
	}

	newEvent := model.Event{
		Severity:    severity,
		User:        username,
		Time:        time.Now().UTC(),
		Message:     msg,
		JobMessages: jobMessages,
	}

	h.Events.StoreEvents(newEvent)
}
