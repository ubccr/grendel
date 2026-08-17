// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd/shared"
	"golang.org/x/sync/errgroup"
)

const defaultShutdownTimeout = 10 * time.Second

func shutdownTimeout() time.Duration {
	d := viper.GetDuration("shutdown_timeout")
	if d <= 0 {
		return defaultShutdownTimeout
	}

	return d
}

// normalize drops the errors a server reports when it was the one asked to stop, those are a clean exit rather than a failure
func normalize(err error) error {
	if errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
		return nil
	}

	return err
}

// run starts every selected service and blocks until they have all stopped. The first service to fail, and a cancelled ctx, both bring down the rest
func run(ctx context.Context, svcs []*Service) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// construct everything up front so a bad address or a missing file fails before any service is listening
	runners := make([]Runner, 0, len(svcs))
	for _, svc := range svcs {
		srv, err := svc.New()
		if err != nil {
			return fmt.Errorf("%s: %w", svc.Name, err)
		}

		runners = append(runners, srv)
	}

	for _, svc := range svcs {
		shared.Log.Infof("Starting %s", svc.Name)
	}

	g, gctx := errgroup.WithContext(ctx)

	for i, svc := range svcs {
		srv := runners[i]

		g.Go(func() error {
			// any exit, a failure or a clean one, stops the other services rather than leaving a half running grendel
			defer cancel()

			serveErr := make(chan error, 1)
			go func() { serveErr <- normalize(srv.Serve()) }()

			select {
			case err := <-serveErr:
				if err != nil {
					return fmt.Errorf("%s: %w", svc.Name, err)
				}

				return nil

			case <-gctx.Done():
				shared.Log.Infof("Shutting down %s...", svc.Name)

				sctx, stop := context.WithTimeout(context.Background(), shutdownTimeout())
				defer stop()

				if err := srv.Shutdown(sctx); err != nil {
					return fmt.Errorf("%s: shutdown: %w", svc.Name, err)
				}

				// the udp servers open their listener inside Serve, so a shutdown that lands before they are listening is missed. Bound the drain rather than hanging the process on it
				select {
				case <-serveErr:
				case <-sctx.Done():
				}

				return nil
			}
		})
	}

	return g.Wait()
}
