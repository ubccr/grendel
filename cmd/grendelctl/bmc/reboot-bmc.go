// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendelctl/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	powerBmcFilter  []string
	powerBmcOptions *shared.TableOptions
	powerBmcColumns = []shared.Column{
		{Name: "Node"},
		{Name: "Status"},
		{Name: "Message"},
	}
	powerBmcCmd = &cobra.Command{
		Use:   "reboot-bmc <nodeset | all>",
		Short: "Reboot the BMC",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			params := client.POSTV1BmcPowerBmcParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			jobs, err := apiclient.API.POSTV1BmcPowerBmc(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(powerBmcColumns, powerBmcFilter).Options(*powerBmcOptions).Empty("-")

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
	powerBmcOptions = shared.RegisterTableFlags(powerBmcCmd)
	powerBmcCmd.Flags().StringSliceVar(&powerBmcFilter, "filter", nil, "filter table columns by name")
	powerBmcCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(powerBmcColumns))

	bmcCmd.AddCommand(powerBmcCmd)
}
