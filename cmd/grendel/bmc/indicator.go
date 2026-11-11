// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	indicatorFilter  []string
	indicatorOptions *shared.TableOptions
	indicatorColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Message"},
	}
	indicatorCmd = &cobra.Command{
		Use:       "indicator {on | off} {nodeset | all}",
		Aliases:   []string{"led", "locate"},
		Short:     "Turn the location indicator on or off",
		Long:      "Turn the location indicator (identify LED) on or off, so the nodes can be found in the rack",
		Args:      cobra.ExactArgs(2),
		ValidArgs: []string{"on", "off"},
		RunE: func(command *cobra.Command, args []string) error {
			if args[0] != "on" && args[0] != "off" {
				return fmt.Errorf("invalid state %q, must be on or off", args[0])
			}

			req := &client.BmcIndicatorBody{State: client.NewOptString(args[0])}

			nodeset, all := selector(args[1])
			params := client.POSTV1BmcIndicatorParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcIndicator(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(indicatorColumns, indicatorFilter).Options(*indicatorOptions).Empty("-")
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
	indicatorOptions = shared.RegisterTableFlags(indicatorCmd)
	indicatorCmd.Flags().StringSliceVar(&indicatorFilter, "filter", nil, "filter table columns by name")
	indicatorCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(indicatorColumns))

	bmcCmd.AddCommand(indicatorCmd)
}
