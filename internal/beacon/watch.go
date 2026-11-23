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

	"github.com/ubccr/grendel/internal/beacon/wire"
)

type Poster interface {
	Post(ctx context.Context, r wire.Report) error
	PostRetry(ctx context.Context, r wire.Report, giveUp time.Duration) error
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
	// Facts go with the first progress report that gets through, so a stuck boot still shows what it is running, and with the final one, which may be the only report there is
	Facts map[string]string
}

// DefaultTimeout is how long a watch waits for the boot when it isn't told. Grendel only takes a report while the node's boot token is valid, an hour from its network boot unless provision.token_ttl says otherwise, and this runs out with time left for the failed report to get in
const DefaultTimeout = 45 * time.Minute

// ErrBootFailed is returned once a failed boot has been reported
var ErrBootFailed = errors.New("boot failed")

// Watch follows systemd until the boot finishes, reporting progress along the way and the result at the end. It returns nil on a clean boot, ErrBootFailed once a failure is reported, and ErrRejected when grendel won't take reports, such as with an expired token
func Watch(ctx context.Context, sys System, poster Poster, opts WatchOptions) error {
	deadline := time.Now().Add(opts.Timeout)
	var lastState string
	var lastJobs int
	var lastSent time.Time
	sentFacts := false

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	for {
		state, err := sys.State(ctx)
		if err != nil {
			logf("failed to read the system state: %s", err)
		}

		var final *wire.Report
		switch state {
		case "":
		case "initializing", "starting":
			jobs, err := sys.Jobs(ctx)
			if err != nil {
				logf("failed to list jobs: %s", err)
			}
			if state != lastState || len(jobs) != lastJobs || time.Since(lastSent) >= opts.Heartbeat {
				r := wire.Report{State: wire.StateBooting, QueuedJobs: len(jobs), BootID: opts.BootID, Message: "system is " + state}
				if !sentFacts {
					r.Facts = opts.Facts
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
			final = &wire.Report{State: wire.StateCompleted, BootID: opts.BootID, Complete: opts.Complete, Message: "system is running"}
		case "degraded":
			units, err := sys.FailedUnits(ctx)
			if err != nil {
				logf("failed to list failed units: %s", err)
			}
			final = &wire.Report{State: wire.StateFailed, BootID: opts.BootID, FailedUnits: units, Message: failedMessage(units)}
		default:
			final = &wire.Report{State: wire.StateFailed, BootID: opts.BootID, Message: "system is " + state}
		}
		lastState = state

		if final == nil && time.Now().After(deadline) {
			jobs, _ := sys.Jobs(ctx)
			units, _ := sys.FailedUnits(ctx)
			final = &wire.Report{State: wire.StateFailed, BootID: opts.BootID, FailedUnits: units, QueuedJobs: len(jobs), Message: timeoutMessage(opts.Timeout, jobs)}
		}

		if final != nil {
			final.Facts = opts.Facts
			logf("reporting %s: %s", final.State, final.Message)
			if err := poster.PostRetry(ctx, *final, opts.FinalRetry); err != nil {
				return err
			}
			if final.State == wire.StateFailed {
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

func failedMessage(units []wire.FailedUnit) string {
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
