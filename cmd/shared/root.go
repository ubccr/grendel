// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package shared

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/version"
)

var (
	cfgFile string
	debug   bool
	verbose bool

	Log = logger.GetLogger("CLI")

	// Root is shared by both binaries. Each command package attaches itself from its init, and the binary's main decides which of those packages get linked in
	Root = &cobra.Command{
		Version: version.Version,
		Long: `Syntax Legend:
  < >   Required parameter
  [ ]   Optional parameter
  ...   Multiple parameters
  |     Mutually Exclusive
`,
	}
)

// Execute names the binary and runs it. The name is not fixed at init because grendeld and grendel share this root
func Execute(use, short string) {
	Root.Use = use
	Root.Short = short

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := Root.ExecuteContext(ctx); err != nil {
		if ctx.Err() != nil {
			os.Exit(130)
		}
		Log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	Root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	Root.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug messages")
	Root.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose messages")

	Root.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		return SetupLogging()
	}
}
