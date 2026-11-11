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
	selClearFilter  []string
	selClearOptions *shared.TableOptions
	selClearColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Message"},
	}
	selClearCmd = &cobra.Command{
		Use:   "clear <nodeset | all>",
		Short: "Clear the system event log on the BMC",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset, all := selector(args[0])
			params := client.POSTV1BmcSelClearParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcSelClear(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(selClearColumns, selClearFilter).Options(*selClearOptions).Empty("-")
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
	selClearOptions = shared.RegisterTableFlags(selClearCmd)
	selClearCmd.Flags().StringSliceVar(&selClearFilter, "filter", nil, "filter table columns by name")
	selClearCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(selClearColumns))

	selCmd.AddCommand(selClearCmd)
}
