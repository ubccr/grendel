// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"fmt"
	"strings"

	"github.com/stmcginnis/gofish/oem/dell"
	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/pkg/model"
)

type Service struct {
	cfg  Config
	opts Options
}

func New(cfg Config) *Service { return &Service{cfg: cfg, opts: cfg.options()} }

func NewFromViper() *Service { return New(ConfigFromViper()) }

type PowerRequest struct {
	Target PowerTarget
	// Reset is ignored when Target is TargetBMC, which is always a graceful restart
	Reset schemas.ResetType
	Boot  schemas.BootSource
}

type PowerTarget string

const (
	TargetHost PowerTarget = "host"
	TargetBMC  PowerTarget = "bmc"
)

type IndicatorRequest struct {
	On bool
}

type ClearJobsRequest struct {
	All bool
	IDs []string
}

type JobList struct {
	Jobs []*schemas.Job
}

type SELRequest struct {
	Limit int
}

type SELLog struct {
	Entries []*schemas.LogEntry
}

// ImportConfigRequest is one import for every target. Body is sent as is apart from ShareParameters.ShareName, which is per node and taken from ShareNames by target name
type ImportConfigRequest struct {
	Body dell.ImportSystemConfigurationBody
	// PortNumber is the share's port, sent only when set. Leave it empty for the scheme's default port, since only iDRAC9 7.00.00 and later accept it
	PortNumber string
	ShareNames map[string]string
}

type InstallFirmwareRequest struct {
	Body dell.InstallFromRepoBody
	// ClearJobs empties the job queue first. Only applies when Body.ApplyUpdate is true
	ClearJobs bool
}

type FirmwareUpdates struct {
	Updates dell.UpdateList
}

// Status reads system information from every target
func (s *Service) Status(ctx context.Context, targets []Target) []NodeResult[model.RedfishSystem] {
	return run(ctx, s, targets, s.statusOne)
}

// Power changes the power state of every target, or of their BMCs
func (s *Service) Power(ctx context.Context, targets []Target, req PowerRequest) []NodeResult[Ack] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*Ack, error) { return s.powerOne(ctx, t, req) })
}

// Indicator turns the location indicator of every target on or off
func (s *Service) Indicator(ctx context.Context, targets []Target, req IndicatorRequest) []NodeResult[Ack] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*Ack, error) { return s.indicatorOne(ctx, t, req) })
}

// Jobs lists the Redfish jobs queued on every target
func (s *Service) Jobs(ctx context.Context, targets []Target) []NodeResult[JobList] {
	return run(ctx, s, targets, s.jobsOne)
}

// ClearJobs removes jobs from every target's queue
func (s *Service) ClearJobs(ctx context.Context, targets []Target, req ClearJobsRequest) []NodeResult[Ack] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*Ack, error) { return s.clearJobsOne(ctx, t, req) })
}

// SEL reads the system event log of every target
func (s *Service) SEL(ctx context.Context, targets []Target, req SELRequest) []NodeResult[SELLog] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*SELLog, error) { return s.selOne(ctx, t, req) })
}

// ClearSEL clears the system event log on every target
func (s *Service) ClearSEL(ctx context.Context, targets []Target) []NodeResult[Ack] {
	return run(ctx, s, targets, s.clearSELOne)
}

// AutoConfigure sets every target's iDRAC to auto configure once on its next boot
func (s *Service) AutoConfigure(ctx context.Context, targets []Target) []NodeResult[Ack] {
	return run(ctx, s, targets, s.autoConfigureOne)
}

// ImportConfig submits a system configuration import job to every target's iDRAC
func (s *Service) ImportConfig(ctx context.Context, targets []Target, req ImportConfigRequest) []NodeResult[Ack] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*Ack, error) { return s.importConfigOne(ctx, t, req) })
}

// FirmwareUpdates lists the updates every target's iDRAC found in its last catalog comparison
func (s *Service) FirmwareUpdates(ctx context.Context, targets []Target) []NodeResult[FirmwareUpdates] {
	return run(ctx, s, targets, s.firmwareUpdatesOne)
}

// InstallFirmware compares every target's firmware against a catalog, and queues the updates when asked to
func (s *Service) InstallFirmware(ctx context.Context, targets []Target, req InstallFirmwareRequest) []NodeResult[Ack] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*Ack, error) { return s.installFirmwareOne(ctx, t, req) })
}

// run applies bmc.run_timeout, so a large nodeset of wedged BMCs still produces a response in bounded time
func run[T any](ctx context.Context, s *Service, targets []Target, fn func(context.Context, Target) (*T, error)) []NodeResult[T] {
	if s.cfg.RunTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.cfg.RunTimeout)
		defer cancel()
	}

	return Run(ctx, targets, s.opts, fn)
}

// done is the payload for operations that report nothing but their outcome
func done() (*Ack, error) { return &Ack{}, nil }

// doneWith reports success along with a note about what was actually done
func doneWith(format string, args ...any) (*Ack, error) {
	return &Ack{Note: fmt.Sprintf(format, args...)}, nil
}

// clearEach attempts every item even after one refuses
func clearEach(noun string, names []string, clear func(i int) error) (*Ack, error) {
	var (
		failed   []string
		firstErr error
	)

	for i, name := range names {
		if err := clear(i); err != nil {
			log.Debugf("failed to clear %s %s: %s", noun, name, err)
			failed = append(failed, name)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	if firstErr != nil {
		return nil, fmt.Errorf("cleared %d of %d %s, failed on %s: %s", len(names)-len(failed), len(names), noun, strings.Join(failed, ", "), errorMessage(firstErr))
	}

	return doneWith("cleared %d %s", len(names), noun)
}
