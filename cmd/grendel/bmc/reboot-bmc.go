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
	powerBmcFilter  []string
	powerBmcOptions *shared.TableOptions
	powerBmcColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Message"},
	}
	powerBmcCmd = &cobra.Command{
		Use:   "reboot-bmc <nodeset | all>",
		Short: "Reboot the BMC",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			// one power route covers both, a bmc target is always a graceful manager restart
			req := &client.BmcPowerBody{Target: client.NewOptString("bmc")}

			nodeset, all := selector(args[0])
			params := client.POSTV1BmcPowerParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcPower(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(powerBmcColumns, powerBmcFilter).Options(*powerBmcOptions).Empty("-")
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
	powerBmcOptions = shared.RegisterTableFlags(powerBmcCmd)
	powerBmcCmd.Flags().StringSliceVar(&powerBmcFilter, "filter", nil, "filter table columns by name")
	powerBmcCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(powerBmcColumns))

	bmcCmd.AddCommand(powerBmcCmd)
}
