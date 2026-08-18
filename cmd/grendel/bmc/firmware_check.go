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
	firmwareCheckFilter  []string
	firmwareCheckOptions *shared.TableOptions
	firmwareCheckColumns = []shared.Column{
		{Name: "Node"},
		{Name: "Status"},
		{Name: "Message"},
		{Name: "Component"},
		{Name: "Current Version"},
		{Name: "Latest Version"},
		{Name: "Reboot Required"},
	}
	firmwareCheckCmd = &cobra.Command{
		Use:   "check <nodeset>",
		Short: "Check for updates on Dell servers",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]

			if nodeset == "all" {
				nodeset = ""
			}

			params := client.GETV1BmcUpgradeDellRepoParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			jobs, err := apiclient.API.GETV1BmcUpgradeDellRepo(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(firmwareCheckColumns, firmwareCheckFilter).Options(*firmwareCheckOptions).Empty("-")

			for _, job := range jobs {
				if len(job.UpdateList.Value) < 1 {
					t.AppendRow(
						job.Name.Value,
						job.Status.Value,
						job.Message.Value,
						"",
						"",
						"",
						"",
					)
					continue
				}

				for _, update := range job.UpdateList.Value {
					t.AppendRow(
						job.Name.Value,
						job.Status.Value,
						job.Message.Value,
						update.Value.DisplayName.Value,
						update.Value.InstalledVersion.Value,
						colorVersion(update.Value.InstalledVersion.Value, update.Value.PackageVersion.Value),
						update.Value.RebootType.Value,
					)
				}
			}

			t.Print()
			return nil
		},
	}
)

func init() {
	firmwareCheckOptions = shared.RegisterTableFlags(firmwareCheckCmd)
	firmwareCheckCmd.Flags().StringSliceVar(&firmwareCheckFilter, "filter", nil, "filter table columns by name")
	firmwareCheckCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(firmwareCheckColumns))

	firmwareCmd.AddCommand(firmwareCheckCmd)
}
