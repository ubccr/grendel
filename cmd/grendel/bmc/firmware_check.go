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
		{Name: "State"},
		{Name: "Message"},
		{Name: "Component"},
		{Name: "Current Version"},
		{Name: "Latest Version"},
		{Name: "Reboot Required"},
	}
	firmwareCheckCmd = &cobra.Command{
		Use:   "check <nodeset | all>",
		Short: "Check for updates on Dell servers",
		Long:  "Check for updates on Dell servers. Lists what the iDRAC found in its last catalog comparison, which bmc firmware upgrade runs without --apply-update",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset, all := selector(args[0])
			params := client.GETV1BmcFirmwareUpdatesParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.GETV1BmcFirmwareUpdates(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(firmwareCheckColumns, firmwareCheckFilter).Options(*firmwareCheckOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				state := colorState(result.Success.Value)

				updates := result.Data.Value.Updates.Value
				if len(updates) == 0 {
					t.AppendRow(result.Node.Value, state, result.Message.Value, "", "", "", "")
					continue
				}

				for _, u := range updates {
					update := u.Value
					t.AppendRow(
						result.Node.Value,
						state,
						result.Message.Value,
						update.DisplayName.Value,
						update.InstalledVersion.Value,
						colorVersion(update.InstalledVersion.Value, update.PackageVersion.Value),
						update.RebootType.Value,
					)
				}
			}
			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

func init() {
	firmwareCheckOptions = shared.RegisterTableFlags(firmwareCheckCmd)
	firmwareCheckCmd.Flags().StringSliceVar(&firmwareCheckFilter, "filter", nil, "filter table columns by name")
	firmwareCheckCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(firmwareCheckColumns))

	firmwareCmd.AddCommand(firmwareCheckCmd)
}
