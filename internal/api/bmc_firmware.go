// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

const (
	// firmwareStageQueued is a node whose update was just queued
	firmwareStageQueued = "queued"
	// firmwareStageVerifying is a node whose jobs all read as finished, which is checked once more before it counts, since an iDRAC adds a job per component a little after the first
	firmwareStageVerifying = "verifying"
)

func init() {
	viper.SetDefault("bmc.firmware_poll_interval", "1m")
}

// firmwareQueued leaves a node whose update was queued as running, so the event stays open for trackFirmware to follow. Its ref is the id of the job the iDRAC gave, which is how its jobs are found again
func firmwareQueued(row *model.EventNode, success bool, ack *redfish.Ack) {
	if !success || ack == nil || ack.Job == "" {
		return
	}
	row.State = model.EventNodeRunning
	row.Stage = firmwareStageQueued
	row.Ref = ack.Job
}

// trackFirmware follows the firmware updates queued on BMCs until their jobs finish, by reading the job queues of the nodes still running in open firmware events. What it knows is all in the events, so it picks up where it was after a restart
func (h *Handler) trackFirmware(ctx context.Context) {
	interval := viper.GetDuration("bmc.firmware_poll_interval")
	if interval <= 0 {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.checkFirmware(ctx)
		}
	}
}

func (h *Handler) checkFirmware(ctx context.Context) {
	open, err := h.DB.Events(model.EventFilter{Kind: model.EventKindFirmware, State: model.EventStateRunning, Limit: model.MaxEventLimit, Nodes: true})
	if err != nil {
		log.Warnf("failed to list running firmware events: %s", err)
		return
	}

	for _, event := range open {
		running := map[string]model.EventNode{}
		names := []string{}
		for _, n := range event.Nodes {
			if n.State == model.EventNodeRunning {
				running[n.Node] = n
				names = append(names, n.Node)
			}
		}
		if len(names) == 0 {
			h.Events.Settle(&event)
			continue
		}

		ns, err := nodeset.NewNodeSet(strings.Join(names, ","))
		if err != nil {
			log.Warnf("failed to read the nodes of firmware event %s: %s", event.ID, err)
			continue
		}
		hosts, err := h.DB.FindHosts(ns)
		if err != nil {
			log.Warnf("failed to load the nodes of firmware event %s: %s", event.ID, err)
			continue
		}

		var changed []model.EventNode
		for _, r := range h.Redfish.Jobs(ctx, redfish.TargetsFromHosts(hosts)) {
			// a BMC that does not answer is most likely restarting to apply the update, so its node is left as it is. One that never comes back is marked stalled after bmc.firmware_timeout
			if !r.Success || r.Data == nil {
				continue
			}
			if row, ok := firmwareProgress(running[r.Node], r.Data.Jobs); ok {
				// written as of now, which restarts how long the node can go without a change
				row.UpdatedAt = time.Time{}
				changed = append(changed, row)
			}
		}
		if len(changed) > 0 {
			h.Events.SetNodes(&event, changed...)
			h.Events.Settle(&event)
		}
	}
}

// firmwareProgress works out where a node's update has got to from its job queue, and whether that differs from its row. The jobs of the update are the one the row's ref names and every job queued after it. Only a change is written, so how long a node has gone without one is what marks it stalled
func firmwareProgress(row model.EventNode, jobs []*schemas.Job) (model.EventNode, bool) {
	first := row.Ref
	next := row
	var active *schemas.Job
	total, finished := 0, 0

	for _, job := range jobs {
		if job == nil || redfish.CompareJobIDs(job.ID, first) < 0 {
			continue
		}
		total++
		switch job.JobState {
		case schemas.ExceptionJobState, schemas.CancelledJobState, schemas.InvalidJobState:
			next.State = model.EventNodeFailed
			next.Stage = strings.ToLower(string(job.JobState))
			next.Message = strings.TrimSpace(fmt.Sprintf("%s %s: %s", job.ID, job.Name, jobMessage(job.Messages)))
			return next, true
		case schemas.CompletedJobState:
			finished++
		default:
			if job.PercentComplete != nil && *job.PercentComplete >= 100 {
				finished++
			} else if active == nil || redfish.CompareJobIDs(job.ID, active.ID) < 0 {
				// the oldest job not yet done is the one the update is on
				active = job
			}
		}
	}

	switch {
	case total == 0:
		// the queue no longer has the job, as after it was cleared, so there is nothing to follow
		return row, false
	case active != nil:
		next.Stage = strings.ToLower(string(active.JobState))
	case row.Stage == firmwareStageVerifying:
		next.State = model.EventNodeSucceeded
		next.Message = fmt.Sprintf("%d jobs completed", finished)
	default:
		next.Stage = firmwareStageVerifying
	}

	return next, next.State != row.State || next.Stage != row.Stage
}
