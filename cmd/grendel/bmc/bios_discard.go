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
	biosDiscardFilter  []string
	biosDiscardOptions *shared.TableOptions
	biosDiscardColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Message"},
	}
	biosDiscardCmd = &cobra.Command{
		Use:   "discard <nodeset | all>",
		Short: "Discard staged BIOS changes",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset, all := selector(args[0])
			params := client.POSTV1BmcBiosDiscardParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcBiosDiscard(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(biosDiscardColumns, biosDiscardFilter).Options(*biosDiscardOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				t.AppendRow(result.Node.Value, colorState(result.Success.Value), result.Message.Value)
			}
			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

func init() {
	biosDiscardOptions = shared.RegisterTableFlags(biosDiscardCmd)
	biosDiscardCmd.Flags().StringSliceVar(&biosDiscardFilter, "filter", nil, "filter table columns by name")
	biosDiscardCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(biosDiscardColumns))

	biosCmd.AddCommand(biosDiscardCmd)
}
