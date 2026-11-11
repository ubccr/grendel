// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type Options struct {
	Fanout      int
	NodeTimeout time.Duration
	RateLimit   *rate.Limiter
}

// Run contacts every target and returns exactly one result per target, in input order, whatever happens
func Run[T any](
	ctx context.Context,
	targets []Target,
	opts Options,
	fn func(context.Context, Target) (*T, error),
) []NodeResult[T] {
	results := make([]NodeResult[T], len(targets))

	fanout := opts.Fanout
	if fanout < 1 {
		fanout = 1
	}

	var (
		wg  sync.WaitGroup
		sem = make(chan struct{}, fanout)
	)

	// populate results even if request was cancelled
	cancelRest := func(from int) {
		for j := from; j < len(targets); j++ {
			results[j] = NodeResult[T]{
				Node:    targets[j].Name,
				Message: "canceled before this node was contacted",
			}
		}
	}

dispatch:
	for i := range targets {
		if opts.RateLimit != nil {
			if err := opts.RateLimit.Wait(ctx); err != nil {
				cancelRest(i)
				break
			}
		}

		select {
		case <-ctx.Done():
			cancelRest(i)
			break dispatch
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(idx int, t Target) {
			defer wg.Done()
			defer func() { <-sem }()

			results[idx] = runOne(ctx, t, opts, fn)
		}(i, targets[i])
	}

	wg.Wait()
	return results
}

// runOne executes the function against a bmc and will always return a response
func runOne[T any](
	ctx context.Context,
	t Target,
	opts Options,
	fn func(context.Context, Target) (*T, error),
) (res NodeResult[T]) {
	started := time.Now()
	res = NodeResult[T]{
		Node:    t.Name,
		Success: false,
		Data:    nil,
	}

	// goroutine panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.WithFields(logrus.Fields{
				"node":  t.Name,
				"panic": r,
				"stack": string(debug.Stack()),
			}).Error("recovered panic during redfish operation")

			res = NodeResult[T]{
				Node:    t.Name,
				Message: fmt.Sprintf("internal error: %v", r),
			}
		}
		res.Duration = time.Since(started)
	}()

	if opts.NodeTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.NodeTimeout)
		defer cancel()
	}

	data, err := fn(ctx, t)
	if err != nil {
		log.Debugf("%s failed: %s", t.Name, err)
		res.Message = errorMessage(err)
		return res
	}

	// neither data nor an error is a bug in the operation, so it is not reported as a success
	if data == nil {
		res.Message = "internal error: operation returned no result"
		return res
	}

	res.Success = true
	res.Data = data
	if n, ok := any(*data).(noter); ok && res.Message == "" {
		res.Message = n.note()
	}
	return res
}
