// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testTargets(n int) []Target {
	targets := make([]Target, n)
	for i := range targets {
		targets[i] = Target{Name: fmt.Sprintf("node-%02d", i), Addr: fmt.Sprintf("10.0.0.%d", i+1)}
	}
	return targets
}

func ok(v string) func(context.Context, Target) (*string, error) {
	return func(context.Context, Target) (*string, error) { return &v, nil }
}

// TestRunReturnsRowPerTarget is the package's central invariant: a mixed run returns every target it was given
func TestRunReturnsRowPerTarget(t *testing.T) {
	targets := testTargets(12)

	results := Run(context.Background(), targets, Options{Fanout: 4},
		func(_ context.Context, tg Target) (*string, error) {
			switch tg.Name {
			case "node-01", "node-05":
				return nil, errors.New("boom")
			case "node-07":
				panic("exploded")
			case "node-09":
				return nil, nil // neither data nor error, a bug in the operation
			}
			v := "fine"
			return &v, nil
		})

	require.Len(t, results, len(targets))

	for i, r := range results {
		assert.Equal(t, targets[i].Name, r.Node, "results must stay in input order")
		if r.Success {
			assert.NotNil(t, r.Data, "success must carry data")
		} else {
			assert.Nil(t, r.Data, "only success may carry data")
			assert.NotEmpty(t, r.Message, "a failed node has to say why")
		}
	}
}

func TestRunPanicIsolation(t *testing.T) {
	targets := testTargets(5)

	results := Run(context.Background(), targets, Options{Fanout: 5},
		func(_ context.Context, tg Target) (*string, error) {
			if tg.Name == "node-02" {
				var systems []string
				return &systems[0], nil
			}
			v := "fine"
			return &v, nil
		})

	require.Len(t, results, 5)
	assert.False(t, results[2].Success)
	assert.Contains(t, results[2].Message, "internal error")
	assert.True(t, results[4].Success, "the other nodes must still complete")
}

func TestRunRespectsFanout(t *testing.T) {
	var (
		inFlight atomic.Int32
		peak     atomic.Int32
	)

	Run(context.Background(), testTargets(30), Options{Fanout: 4},
		func(context.Context, Target) (*string, error) {
			cur := inFlight.Add(1)
			defer inFlight.Add(-1)

			for {
				old := peak.Load()
				if cur <= old || peak.CompareAndSwap(old, cur) {
					break
				}
			}

			time.Sleep(5 * time.Millisecond)
			v := "fine"
			return &v, nil
		})

	assert.LessOrEqual(t, int(peak.Load()), 4)
	assert.Greater(t, int(peak.Load()), 1, "the run should actually be concurrent")
}

func TestRunZeroFanout(t *testing.T) {
	results := Run(context.Background(), testTargets(3), Options{Fanout: 0}, ok("fine"))
	assert.Len(t, results, 3)
}

func TestRunCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	targets := testTargets(20)

	var started atomic.Int32
	begin := time.Now()

	results := Run(ctx, targets, Options{Fanout: 1},
		func(nodeCtx context.Context, _ Target) (*string, error) {
			if started.Add(1) == 3 {
				cancel()
			}
			select {
			case <-nodeCtx.Done():
				return nil, nodeCtx.Err()
			case <-time.After(10 * time.Millisecond):
				v := "fine"
				return &v, nil
			}
		})

	elapsed := time.Since(begin)

	require.Len(t, results, len(targets), "a cancelled run still reports every node")
	assert.Less(t, elapsed, 500*time.Millisecond, "cancellation must not grind through the remaining nodes")

	var canceled int
	for _, r := range results {
		if r.Message == "canceled before this node was contacted" {
			canceled++
		}
	}
	assert.Greater(t, canceled, 0, "undispatched nodes must say they were never contacted")
}

func TestRunNodeTimeout(t *testing.T) {
	results := Run(context.Background(), testTargets(2), Options{Fanout: 2, NodeTimeout: 30 * time.Millisecond},
		func(nodeCtx context.Context, _ Target) (*string, error) {
			<-nodeCtx.Done()
			return nil, nodeCtx.Err()
		})

	require.Len(t, results, 2)
	for _, r := range results {
		assert.False(t, r.Success)
		assert.Contains(t, r.Message, "deadline")
		assert.NotZero(t, r.Duration)
	}
}

func TestRunNoTargets(t *testing.T) {
	assert.Empty(t, Run(context.Background(), nil, Options{Fanout: 4}, ok("fine")))
}
