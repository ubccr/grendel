// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package beacon

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Poster sends a report, the way Client does
type Poster interface {
	Post(ctx context.Context, r Report) error
	PostRetry(ctx context.Context, r Report, giveUp time.Duration) error
}

type WatchOptions struct {
	// Interval is how often systemd is polled
	Interval time.Duration
	// Heartbeat is how often a booting report is repeated when nothing has changed, so grendel can tell a slow boot from a dead beacon
	Heartbeat time.Duration
	// Timeout ends the watch with a failed report when the boot hasn't finished by then
	Timeout time.Duration
	// Complete asks grendel to turn provisioning off for the node on a clean finish
	Complete bool
	BootID   string
	// FinalRetry is how long the last report is retried, since it is the one that matters
	FinalRetry time.Duration
	// Facts gathers the node's facts. It runs for the first progress report, so a stuck boot still shows what it is running, and again for the final one, once the boot timings exist
	Facts func(ctx context.Context) map[string]string
}

// ErrBootFailed is returned once a failed boot has been reported
var ErrBootFailed = errors.New("boot failed")

// Watch follows systemd until the boot finishes, reporting progress along the way and the result at the end. It returns nil on a clean boot, ErrBootFailed once a failure is reported, and ErrRejected when grendel won't take reports, such as with an expired token
func Watch(ctx context.Context, sys System, poster Poster, opts WatchOptions) error {
	deadline := time.Now().Add(opts.Timeout)
	var lastState string
	var lastJobs int
	var lastSent time.Time
	sentFacts := false
	facts := func() map[string]string {
		if opts.Facts == nil {
			return nil
		}
		return opts.Facts(ctx)
	}

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	for {
		state, err := sys.State(ctx)
		if err != nil {
			logf("failed to read the system state: %s", err)
		}

		var final *Report
		switch state {
		case "":
		case "initializing", "starting":
			jobs, err := sys.Jobs(ctx)
			if err != nil {
				logf("failed to list jobs: %s", err)
			}
			if state != lastState || len(jobs) != lastJobs || time.Since(lastSent) >= opts.Heartbeat {
				r := Report{State: StateBooting, QueuedJobs: len(jobs), BootID: opts.BootID, Message: "system is " + state}
				if !sentFacts {
					r.Facts = facts()
				}
				// a progress report that doesn't get through is replaced by the next one, only a rejection is worth stopping for
				if err := poster.Post(ctx, r); errors.Is(err, ErrRejected) {
					return err
				} else if err != nil {
					logf("progress report failed: %s", err)
				} else {
					lastSent = time.Now()
					sentFacts = sentFacts || r.Facts != nil
				}
				lastJobs = len(jobs)
			}
		case "running":
			final = &Report{State: StateCompleted, BootID: opts.BootID, Complete: opts.Complete, Message: "system is running"}
		case "degraded":
			units, err := sys.FailedUnits(ctx)
			if err != nil {
				logf("failed to list failed units: %s", err)
			}
			final = &Report{State: StateFailed, BootID: opts.BootID, FailedUnits: units, Message: failedMessage(units)}
		default:
			final = &Report{State: StateFailed, BootID: opts.BootID, Message: "system is " + state}
		}
		lastState = state

		if final == nil && time.Now().After(deadline) {
			jobs, _ := sys.Jobs(ctx)
			units, _ := sys.FailedUnits(ctx)
			final = &Report{State: StateFailed, BootID: opts.BootID, FailedUnits: units, QueuedJobs: len(jobs), Message: timeoutMessage(opts.Timeout, jobs)}
		}

		if final != nil {
			final.Facts = facts()
			logf("reporting %s: %s", final.State, final.Message)
			if err := poster.PostRetry(ctx, *final, opts.FinalRetry); err != nil {
				return err
			}
			if final.State == StateFailed {
				return ErrBootFailed
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func failedMessage(units []FailedUnit) string {
	if len(units) == 0 {
		return "system is degraded"
	}
	names := make([]string, 0, len(units))
	for _, u := range units {
		names = append(names, u.Name)
	}
	return fmt.Sprintf("%d failed: %s", len(units), listNames(names))
}

func timeoutMessage(timeout time.Duration, jobs []string) string {
	msg := fmt.Sprintf("boot did not finish within %s", timeout)
	if len(jobs) > 0 {
		msg += ", still waiting on " + listNames(jobs)
	}
	return msg
}

// listNames keeps a message readable when hundreds of units are involved
func listNames(names []string) string {
	names = slices.Clone(names)
	slices.Sort(names)
	const show = 10
	if len(names) > show {
		return strings.Join(names[:show], ", ") + fmt.Sprintf(" and %d more", len(names)-show)
	}
	return strings.Join(names, ", ")
}
