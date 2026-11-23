// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package beacon

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/ubccr/grendel/internal/beacon/wire"
)

// System is what the beacon needs from the init system, so Watch can be tested without systemd
type System interface {
	// State is systemd's SystemState: initializing, starting, running, degraded, maintenance or stopping
	State(ctx context.Context) (string, error)
	FailedUnits(ctx context.Context) ([]wire.FailedUnit, error)
	// Jobs lists the units with a job still queued
	Jobs(ctx context.Context) ([]string, error)
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
	p, err := s.conn.SystemStateContext(ctx)
	if err != nil {
		return "", err
	}
	state, ok := p.Value.Value().(string)
	if !ok {
		return "", fmt.Errorf("systemd's SystemState is a %T, not a string", p.Value.Value())
	}
	return state, nil
}

func (s *systemd) FailedUnits(ctx context.Context) ([]wire.FailedUnit, error) {
	units, err := s.conn.ListUnitsFilteredContext(ctx, []string{"failed"})
	if err != nil {
		return nil, err
	}

	failed := make([]wire.FailedUnit, 0, len(units))
	for _, u := range units {
		f := wire.FailedUnit{Name: u.Name, Description: u.Description, ActiveState: u.ActiveState, SubState: u.SubState}
		// Result says why a service failed (exit-code, timeout, signal...). Other unit types report it under their own interface, and the sub state already says enough for them
		if strings.HasSuffix(u.Name, ".service") {
			if p, err := s.conn.GetServicePropertyContext(ctx, u.Name, "Result"); err == nil {
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

func (s *systemd) Close() {
	s.conn.Close()
}
