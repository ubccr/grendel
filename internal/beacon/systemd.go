// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package beacon

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"
)

// System is what the beacon needs from the init system, so Watch can be tested without systemd
type System interface {
	// State is systemd's SystemState: initializing, starting, running, degraded, maintenance or stopping
	State(ctx context.Context) (string, error)
	// FailedUnits lists the units in the failed state
	FailedUnits(ctx context.Context) ([]FailedUnit, error)
	// Jobs lists the units with a job still queued
	Jobs(ctx context.Context) ([]string, error)
	// BootTimes returns the manager's monotonic boot timestamps in microseconds, keyed by property name such as FinishTimestampMonotonic, which is 0 until the boot has finished
	BootTimes(ctx context.Context) (map[string]uint64, error)
	Close()
}

type systemd struct {
	conn *dbus.Conn
}

// Systemd connects to systemd over the system bus, or straight to /run/systemd/private when there is no dbus daemon, so nothing beyond systemd has to be installed
func Systemd(ctx context.Context) (System, error) {
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &systemd{conn: conn}, nil
}

func (s *systemd) State(ctx context.Context) (string, error) {
	v, err := s.conn.GetManagerProperty("SystemState")
	if err != nil {
		return "", err
	}
	// the property comes back as a gvariant string, quoted
	if unquoted, err := strconv.Unquote(v); err == nil {
		return unquoted, nil
	}
	return strings.Trim(v, `"'`), nil
}

func (s *systemd) FailedUnits(ctx context.Context) ([]FailedUnit, error) {
	units, err := s.conn.ListUnitsFilteredContext(ctx, []string{"failed"})
	if err != nil {
		return nil, err
	}

	failed := make([]FailedUnit, 0, len(units))
	for _, u := range units {
		f := FailedUnit{Name: u.Name, Description: u.Description, ActiveState: u.ActiveState, SubState: u.SubState}
		// Result says why a service failed (exit-code, timeout, signal...). Other unit types report it under their own interface, and the sub state already says enough for them
		if strings.HasSuffix(u.Name, ".service") {
			if p, err := s.conn.GetUnitTypePropertyContext(ctx, u.Name, "Service", "Result"); err == nil {
				if r, ok := p.Value.Value().(string); ok {
					f.Result = r
				}
			}
		}
		failed = append(failed, f)
	}
	return failed, nil
}

func (s *systemd) Jobs(ctx context.Context) ([]string, error) {
	jobs, err := s.conn.ListJobsContext(ctx)
	if err != nil {
		return nil, err
	}

	units := make([]string, 0, len(jobs))
	for _, j := range jobs {
		units = append(units, j.Unit)
	}
	return units, nil
}

var bootTimestamps = []string{
	"FirmwareTimestampMonotonic",
	"LoaderTimestampMonotonic",
	"InitRDTimestampMonotonic",
	"UserspaceTimestampMonotonic",
	"FinishTimestampMonotonic",
}

func (s *systemd) BootTimes(ctx context.Context) (map[string]uint64, error) {
	out := map[string]uint64{}
	for _, prop := range bootTimestamps {
		v, err := s.conn.GetManagerProperty(prop)
		if err != nil {
			return nil, err
		}
		// a uint64 property comes back as gvariant text, "@t 12345"
		fields := strings.Fields(v)
		if len(fields) == 0 {
			continue
		}
		n, err := strconv.ParseUint(fields[len(fields)-1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", prop, err)
		}
		out[prop] = n
	}
	return out, nil
}

func (s *systemd) Close() {
	s.conn.Close()
}
