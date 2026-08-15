// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	configureImportFilter  []string
	configureImportOptions *cmd.TableOptions
	configureImportColumns = []cmd.Column{
		{Name: "Node"},
		{Name: "Status"},
		{Name: "Message"},
	}

	configureImportCmd = &cobra.Command{
		Use:   "import <nodeset | all> <NoReboot | Graceful | Forced> <filename>",
		Short: "Import BMC configuration",
		Long: `Args:
	shutdownType:	action bmc should take, NoReboot will wait for the node to be rebooted manually before applying. Any other type WILL REBOOT THE NODE.
	filename:	idrac config file relative to grendel template folder.
		`,
		Args: cobra.ExactArgs(3),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}

			req := &client.BmcImportConfigurationRequest{
				ShutdownType: client.NewOptString(args[1]),
				File:         client.NewOptString(args[2]),
			}
			params := client.POSTV1BmcConfigureImportParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			jobs, err := cmd.API.POSTV1BmcConfigureImport(command.Context(), req, params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			t := cmd.NewTable(configureImportColumns, configureImportFilter).Options(*configureImportOptions).Empty("-")

			for _, job := range jobs {
				t.AppendRow(
					job.Host.Value,
					job.Status.Value,
					job.Msg.Value,
				)
			}

			t.Print()
			return nil
		},
	}
)

func init() {
	configureImportOptions = cmd.RegisterTableFlags(configureImportCmd)
	configureImportCmd.Flags().StringSliceVar(&configureImportFilter, "filter", nil, "filter table columns by name")
	configureImportCmd.RegisterFlagCompletionFunc("filter", cmd.ColumnCompletion(configureImportColumns))

	configureCmd.AddCommand(configureImportCmd)
}
