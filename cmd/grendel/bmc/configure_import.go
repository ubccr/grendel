// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	configureImportFilter  []string
	configureImportOptions *shared.TableOptions
	configureImportColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
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
			req := &client.BmcImportConfigBody{
				ShutdownType: client.NewOptString(args[1]),
				File:         client.NewOptString(args[2]),
			}

			nodeset, all := selector(args[0])
			params := client.POSTV1BmcConfigureImportParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcConfigureImport(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(configureImportColumns, configureImportFilter).Options(*configureImportOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				t.AppendRow(
					result.Node.Value,
					colorState(result.Success.Value),
					result.Message.Value,
				)
			}
			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

func init() {
	configureImportOptions = shared.RegisterTableFlags(configureImportCmd)
	configureImportCmd.Flags().StringSliceVar(&configureImportFilter, "filter", nil, "filter table columns by name")
	configureImportCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(configureImportColumns))

	configureCmd.AddCommand(configureImportCmd)
}
