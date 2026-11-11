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
		entries, err := selEntries(l, req.Limit)
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

// selEntries reads one log service's entries, asking for only the first limit of them when there is a limit
func selEntries(l *schemas.LogService, limit int) ([]*schemas.LogEntry, error) {
	if limit > 0 {
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

// selLogServices finds the system event log, in the same places for reading and clearing. Vendors keep it in different places and some expose another copy elsewhere, so the first place with a match wins, which keeps a log from being read or cleared twice. Systems is only fetched when the managers have nothing
func selLogServices(c *conn) ([]*schemas.LogService, error) {
	managers, err := c.service.Managers()
	if err != nil {
		return nil, err
	}

	var managerLogs [][]*schemas.LogService
	for _, m := range managers {
		services, err := m.LogServices()
		if err != nil {
			return nil, err
		}
		managerLogs = append(managerLogs, services)
	}
	if found := pickLogs(managerLogs, isTypedSEL); len(found) > 0 {
		return found, nil
	}

	systems, err := c.service.Systems()
	if err != nil {
		return nil, err
	}

	var systemLogs [][]*schemas.LogService
	for _, sys := range systems {
		services, err := sys.LogServices()
		if err != nil {
			return nil, err
		}
		systemLogs = append(systemLogs, services)
	}
	if found := pickLogs(systemLogs, isTypedSEL); len(found) > 0 {
		return found, nil
	}
	if found := pickLogs(systemLogs, isKnownSystemLog); len(found) > 0 {
		return found, nil
	}

	// a service that declares no type and is alone on its resource has nothing to be confused with
	if found := pickLogs(managerLogs, isLoneUntypedLog); len(found) > 0 {
		return found, nil
	}
	return pickLogs(systemLogs, isLoneUntypedLog), nil
}

// pickLogs applies a rule to the log services of each manager or system, which is the group a lone service is alone in
func pickLogs(groups [][]*schemas.LogService, match func(l *schemas.LogService, group []*schemas.LogService) bool) []*schemas.LogService {
	var found []*schemas.LogService
	for _, group := range groups {
		for _, l := range group {
			if !isDellFaultList(l) && match(l, group) {
				found = append(found, l)
			}
		}
	}
	return found
}

func isTypedSEL(l *schemas.LogService, _ []*schemas.LogService) bool {
	return l.LogEntryType == schemas.SELLogEntryTypes
}

func isKnownSystemLog(l *schemas.LogService, _ []*schemas.LogService) bool {
	return isKnownSystemLogID(l.ID)
}

func isLoneUntypedLog(l *schemas.LogService, group []*schemas.LogService) bool {
	return l.LogEntryType == "" && len(group) == 1
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
