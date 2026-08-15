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
	configureAutoFilter  []string
	configureAutoOptions *cmd.TableOptions
	configureAutoColumns = []cmd.Column{
		{Name: "Node"},
		{Name: "Status"},
		{Name: "Message"},
	}
	configureAutoCmd = &cobra.Command{
		Use:   "auto <nodeset | all>",
		Short: "Set iDRAC to Auto configure",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}

			params := client.POSTV1BmcConfigureAutoParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := cmd.API.POSTV1BmcConfigureAuto(command.Context(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}
			t := cmd.NewTable(configureAutoColumns, configureAutoFilter).Options(*configureAutoOptions).Empty("-")

			for _, jobMessage := range res {
				t.AppendRow(
					jobMessage.Host.Value,
					jobMessage.Status.Value,
					jobMessage.Msg.Value,
				)
			}

			t.Print()
			return nil
		},
	}
)

func init() {
	configureAutoOptions = cmd.RegisterTableFlags(configureAutoCmd)
	configureAutoCmd.Flags().StringSliceVar(&configureAutoFilter, "filter", nil, "filter table columns by name")
	configureAutoCmd.RegisterFlagCompletionFunc("filter", cmd.ColumnCompletion(configureAutoColumns))

	configureCmd.AddCommand(configureAutoCmd)
}
