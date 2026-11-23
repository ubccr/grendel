// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/stmcginnis/gofish/schemas"
)

// selOne reads the system event log. A limit is sent as $top, which returns the first n members in whatever order the BMC lists them. Redfish defines no $orderby and no order for collection members, so on a BMC that does not list newest first, the limited log is not the newest n entries
func (s *Service) selOne(ctx context.Context, t Target, req SELRequest) (*SELLog, error) {
	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	services, err := selLogServices(c)
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, errors.New("bmc exposes no system event log")
	}

	var out []*schemas.LogEntry
	for _, l := range services {
		entries, err := selEntries(l, req.Limit, c.expandsCollections())
		if err != nil {
			return nil, err
		}
		out = append(out, entries...)
	}

	sort.SliceStable(out, func(i, j int) bool { return entryTime(out[i]).After(entryTime(out[j])) })

	if req.Limit > 0 && len(out) > req.Limit {
		out = out[:req.Limit]
	}

	return &SELLog{Entries: out}, nil
}

// selEntries reads one log service's entries, asking for only the first limit of them when there is a limit. A connection that expands collections reads the whole log instead, because gofish appends ?$expand after the ?$top and the uri would carry two query strings
func selEntries(l *schemas.LogService, limit int, expands bool) ([]*schemas.LogEntry, error) {
	if limit > 0 && !expands {
		return l.FilteredEntries(schemas.WithTop(limit))
	}
	return l.Entries()
}

func (s *Service) clearSELOne(ctx context.Context, t Target) (*Ack, error) {
	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	services, err := selLogServices(c)
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, errors.New("bmc exposes no system event log to clear")
	}

	names := make([]string, 0, len(services))
	for _, l := range services {
		names = append(names, l.ID)
	}

	return clearEach("log service(s)", names, func(i int) error {
		_, err := services[i].ClearLog("")
		return err
	})
}

// selLogServices finds the system event log, in the same places for reading and clearing: the log services typed SEL on the managers, else those on the systems, else a manager log that goes by a known id, or declares no type and is alone. The first place with a match wins, which keeps a copy exposed in two places from being read or cleared twice, and Systems is only fetched when the managers type none
func selLogServices(c *conn) ([]*schemas.LogService, error) {
	managers, err := c.service.Managers()
	if err != nil {
		return nil, err
	}

	var typed, fallback []*schemas.LogService
	for _, m := range managers {
		services, err := m.LogServices()
		if err != nil {
			return nil, err
		}
		for _, l := range services {
			switch {
			case isDellFaultList(l):
			case l.LogEntryType == schemas.SELLogEntryTypes:
				typed = append(typed, l)
			case isSELByID(l.ID), l.LogEntryType == "" && len(services) == 1:
				fallback = append(fallback, l)
			}
		}
	}
	if len(typed) > 0 {
		return typed, nil
	}

	systems, err := c.service.Systems()
	if err != nil {
		return nil, err
	}
	for _, sys := range systems {
		services, err := sys.LogServices()
		if err != nil {
			return nil, err
		}
		for _, l := range services {
			if l.LogEntryType == schemas.SELLogEntryTypes && !isDellFaultList(l) {
				typed = append(typed, l)
			}
		}
	}
	if len(typed) > 0 {
		return typed, nil
	}
	return fallback, nil
}

// entryTime reads an entry's timestamp. An entry carrying none sorts as the oldest rather than to the top of the log, where it would crowd out the entries someone is actually looking for
func entryTime(e *schemas.LogEntry) time.Time {
	if e == nil {
		return time.Time{}
	}

	// Created is the Redfish field, but some services populate only EventTimestamp
	for _, raw := range []string{e.Created, e.EventTimestamp} {
		if raw == "" {
			continue
		}
		if ts, err := time.Parse(time.RFC3339, raw); err == nil {
			return ts
		}
	}

	return time.Time{}
}
