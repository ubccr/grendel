// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"encoding/json"
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
	// firmwareStageScheduled is a node whose update is staged and waits for the node's next reboot, which the install was asked not to do
	firmwareStageScheduled = "scheduled"
	// firmwareStageGone is a node whose update's jobs are no longer in its BMC's queue, as after the queue was cleared, so how it went is not known
	firmwareStageGone = "gone"

	// firmwareFollowFor is how long after an update was queued its nodes are followed. A node still running then is stalled, which keeps an update that never finishes from holding its event open, and a stalled node is checked until then in case its BMC comes back with the outcome
	firmwareFollowFor = 24 * time.Hour

	// firmwareDetailReboot is the detail of a firmware event saying whether the install was let reboot the nodes, "true" or "false"
	firmwareDetailReboot = "reboot"
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

// jobReader reads the job queues of BMCs, which is redfish.Service.Jobs outside of tests
type jobReader func(context.Context, []redfish.Target) []redfish.NodeResult[redfish.JobList]

// trackFirmware follows the firmware updates queued on BMCs until their jobs finish, by reading the job queues of the nodes in recent firmware events. What it knows is all in the events, so it picks up where it was after a restart
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
			h.checkFirmware(ctx, time.Now().UTC(), h.Redfish.Jobs)
		}
	}
}

// checkFirmware reads the job queues of the nodes firmware events still follow: the running nodes of running events, and the stalled nodes of events queued within firmwareFollowFor, whose BMC may have come back
//
// A running node is written every time its BMC answers, changed or not, so the time it was last written is when its BMC last answered, and Tidy stalls it once that is longer ago than bmc.firmware_timeout. A BMC that does not answer is most likely restarting to apply the update, and its node is left as it is
func (h *Handler) checkFirmware(ctx context.Context, now time.Time, read jobReader) {
	recent, err := h.DB.Events(model.EventFilter{Kind: model.EventKindFirmware, Limit: model.MaxEventLimit, Nodes: true})
	if err != nil {
		log.Warnf("failed to list firmware events: %s", err)
		return
	}

	for _, event := range recent {
		running := event.State == model.EventStateRunning
		expired := now.Sub(event.StartedAt) > firmwareFollowFor
		if !running && expired {
			continue
		}

		followed := map[string]model.EventNode{}
		names := []string{}
		for _, n := range event.Nodes {
			if n.Ref != "" && (n.State == model.EventNodeRunning || n.State == model.EventNodeStalled) {
				followed[n.Node] = n
				names = append(names, n.Node)
			}
		}
		if len(names) == 0 {
			if running {
				h.Events.Settle(&event)
			}
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

		waitsForReboot := event.Detail[firmwareDetailReboot] == "false"
		var rows []model.EventNode
		for _, r := range read(ctx, redfish.TargetsFromHosts(hosts)) {
			if !r.Success || r.Data == nil {
				continue
			}
			row := followed[r.Node]
			next, changed := firmwareProgress(row, r.Data.Jobs, waitsForReboot)
			if next.State == model.EventNodeRunning && expired {
				next.State = model.EventNodeStalled
				next.Message = fmt.Sprintf("not finished %s after it was queued, at %s", firmwareFollowFor, next.Stage)
				changed = true
			}
			// a running node is written whenever its BMC answers, which is what keeps it from being stalled. A stalled one is only written when it has something new, its outcome most of all
			if changed || row.State == model.EventNodeRunning {
				next.UpdatedAt = time.Time{}
				rows = append(rows, next)
			}
		}
		if len(rows) > 0 {
			h.Events.SetNodes(&event, rows...)
			h.Events.Settle(&event)
		}
	}
}

// firmwareProgress works out where a node's update has got to from its job queue, and whether that differs from its row. The jobs of the update are the one the row's ref names and every job queued after it. A node that is stalled stays stalled until its update has an outcome. waitsForReboot is an update the install was asked not to reboot for, whose node is done with once all that is left of it waits for the next reboot
func firmwareProgress(row model.EventNode, jobs []*schemas.Job, waitsForReboot bool) (model.EventNode, bool) {
	first := row.Ref
	next := row
	var active *schemas.Job
	total, finished, scheduled := 0, 0, 0

	for _, job := range jobs {
		if job == nil || redfish.CompareJobIDs(job.ID, first) < 0 {
			continue
		}
		total++
		dellState := dellJobState(job)
		switch {
		case job.JobState == schemas.ExceptionJobState || job.JobState == schemas.CancelledJobState || job.JobState == schemas.InvalidJobState ||
			dellState == "Failed" || dellState == "CompletedWithErrors":
			next.State = model.EventNodeFailed
			next.Stage = strings.ToLower(string(job.JobState))
			if dellState != "" {
				next.Stage = strings.ToLower(dellState)
			}
			next.Message = strings.TrimSpace(fmt.Sprintf("%s %s: %s", job.ID, job.Name, jobMessage(job.Messages)))
			return next, true
		case job.JobState == schemas.CompletedJobState, job.PercentComplete != nil && *job.PercentComplete >= 100:
			finished++
		default:
			if dellState == "Scheduled" {
				scheduled++
			}
			// the oldest job not yet done is the one the update is on
			if active == nil || redfish.CompareJobIDs(job.ID, active.ID) < 0 {
				active = job
			}
		}
	}

	switch {
	case total == 0:
		// the queue no longer has the job, as after it was cleared, so how the update went cannot be known
		next.State = model.EventNodeStalled
		next.Stage = firmwareStageGone
		next.Message = fmt.Sprintf("job %s is no longer in the BMC's job queue", first)
	case active != nil && waitsForReboot && scheduled == total-finished:
		next.State = model.EventNodeSucceeded
		next.Stage = firmwareStageScheduled
		next.Message = fmt.Sprintf("%d updates scheduled for the node's next reboot", scheduled)
	case active != nil:
		next.Stage = strings.ToLower(string(active.JobState))
	case row.Stage == firmwareStageVerifying:
		next.State = model.EventNodeSucceeded
		next.Message = fmt.Sprintf("%d jobs completed", finished)
	default:
		next.Stage = firmwareStageVerifying
	}

	return next, next.State != row.State || next.Stage != row.Stage || next.Message != row.Message
}

// dellJobState is the job state an iDRAC gives a job under its Oem section, such as Scheduled for a job waiting for the next reboot, which the Redfish job states have no word for. It is empty for any other BMC
func dellJobState(job *schemas.Job) string {
	if len(job.OEM) == 0 {
		return ""
	}
	var oem struct {
		Dell struct {
			JobState string
		}
	}
	if err := json.Unmarshal(job.OEM, &oem); err != nil {
		return ""
	}
	return oem.Dell.JobState
}
