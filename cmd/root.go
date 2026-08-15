// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/api"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	cfgFile string
	debug   bool
	verbose bool
	API     *client.Client

	Log  = logger.GetLogger("CLI")
	Root = &cobra.Command{
		Use:     "grendel",
		Version: api.Version,
		Short:   "Bare Metal Provisioning for HPC",
		Long: `Syntax Legend:
  < >   Required parameter
  [ ]   Optional parameter
  ...   Multiple parameters
  |     Mutually Exclusive
`,
	}
)

func Execute() {
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
	Root.PersistentFlags().String("endpoint", "grendel-api.socket", "Grendel API endpoint")
	viper.BindPFlag("client.api_endpoint", Root.PersistentFlags().Lookup("endpoint"))

	Root.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		err := SetupLogging()
		if err != nil {
			return err
		}

		// Note that cobra's __complete command sets DisableFlagParsing, completions resolve the endpoint from the config file and env only, and an inline --endpoint or --config is ignored.
		gc, err := NewOgenClient()
		if err != nil {
			return err
		}
		API = gc

		return nil
	}
}
