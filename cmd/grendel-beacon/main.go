// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/internal/beacon"
	"github.com/ubccr/grendel/internal/version"
)

var (
	statusURL string
	insecure  bool
	complete  bool
	factsDir  string
)

func main() {
	root := &cobra.Command{
		Use:           "grendel-beacon",
		Short:         "Report a node's provision status back to grendel",
		Version:       version.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&statusURL, "url", os.Getenv("GRENDEL_STATUS_URL"), "status URL of this node, {{ $.endpoints.StatusURL }} in a provision template (env GRENDEL_STATUS_URL)")
	root.PersistentFlags().BoolVar(&insecure, "insecure", os.Getenv("GRENDEL_INSECURE") == "true", "skip TLS certificate verification (env GRENDEL_INSECURE=true)")
	root.PersistentFlags().BoolVar(&complete, "complete", false, "on a clean finish also turn provisioning off for this node, as the complete endpoint does")

	root.PersistentFlags().StringVar(&factsDir, "facts-dir", beacon.DefaultFactsDir, "drop-in directory of site facts: executables printing key=value lines, or plain files of them, each prefixed with its file name")

	root.AddCommand(watchCmd(), reportCmd(), factsCmd())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := root.ExecuteContext(ctx)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, "grendel-beacon:", err)
		os.Exit(1)
	}
}

func client() (*beacon.Client, error) {
	if statusURL == "" {
		return nil, errors.New("no status url, set --url or GRENDEL_STATUS_URL")
	}
	return beacon.NewClient(statusURL, insecure), nil
}

func watchCmd() *cobra.Command {
	opts := beacon.WatchOptions{Heartbeat: time.Minute, FinalRetry: 5 * time.Minute}

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Follow systemd until the boot finishes and report the result",

		Args: cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			c, err := client()
			if err != nil {
				return err
			}

			sys, err := beacon.Systemd(command.Context())
			if err != nil {
				return fmt.Errorf("failed to connect to systemd: %w", err)
			}
			defer sys.Close()

			opts.Complete = complete
			opts.BootID = beacon.BootID()
			opts.Facts = func(ctx context.Context) map[string]string {
				return beacon.Collect(ctx, factSource(), sys)
			}

			err = beacon.Watch(command.Context(), sys, c, opts)
			switch {
			case err == nil, errors.Is(err, beacon.ErrBootFailed):
				return nil
			case errors.Is(err, beacon.ErrRejected):
				fmt.Fprintln(os.Stderr, "grendel-beacon: not reporting:", err)
				return nil
			}
			return err
		},
	}
	cmd.Flags().DurationVar(&opts.Interval, "interval", 10*time.Second, "how often to check systemd")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", time.Hour, "report the boot as failed when it hasn't finished by then")
	return cmd
}

func factSource() beacon.FactSource {
	src := beacon.DefaultFacts
	src.DropIns = factsDir
	return src
}

// systemdIfAvailable is for commands that can do without systemd, such as a report from an installer that may not run it
func systemdIfAvailable(ctx context.Context) beacon.System {
	sys, err := beacon.Systemd(ctx)
	if err != nil {
		return nil
	}
	return sys
}

func factsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "facts",
		Short: "Print the facts a report would carry, to check a drop-in",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			sys := systemdIfAvailable(command.Context())
			if sys != nil {
				defer sys.Close()
			}
			facts := beacon.Collect(command.Context(), factSource(), sys)
			keys := make([]string, 0, len(facts))
			for k := range facts {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("%s=%s\n", k, facts[k])
			}
			return nil
		},
	}
}

func reportCmd() *cobra.Command {
	var r beacon.Report
	var extra []string
	var collect bool

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Send one status report, from a script or a kickstart %post",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			switch r.State {
			case beacon.StateBooting, beacon.StateFailed, beacon.StateCompleted:
			default:
				return fmt.Errorf("--state must be one of %s, %s, %s", beacon.StateBooting, beacon.StateFailed, beacon.StateCompleted)
			}

			c, err := client()
			if err != nil {
				return err
			}

			// built in and drop-in facts only on request, since an installer's %post would report the installer's os-release rather than the installed OS's
			r.Facts = map[string]string{}
			if collect {
				sys := systemdIfAvailable(command.Context())
				if sys != nil {
					defer sys.Close()
				}
				r.Facts = beacon.Collect(command.Context(), factSource(), sys)
			}
			for _, f := range extra {
				k, v, ok := strings.Cut(f, "=")
				if !ok || k == "" {
					return fmt.Errorf("--fact %q is not key=value", f)
				}
				r.Facts[k] = v
			}

			// no boot id, since a report from an installer's %post comes from a different boot than the installed OS, whose watch has to be able to report after it
			r.Complete = complete
			return c.PostRetry(command.Context(), r, time.Minute)
		},
	}
	cmd.Flags().StringVar(&r.State, "state", "", "booting, failed or completed")
	cmd.Flags().StringVar(&r.Message, "message", "", "shown with the status in grendel")
	cmd.Flags().StringArrayVar(&extra, "fact", nil, "a fact to report, key=value, repeatable. Keys are lowercase with dots, such as site.raid_layout")
	cmd.Flags().BoolVar(&collect, "collect", false, "also report the built in facts (os-release, kernel, boot timings) and the drop-ins")
	return cmd
}
