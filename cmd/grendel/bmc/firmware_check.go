// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	firmwareCheckTable   *shared.TableFlags
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
			params := client.GETV1BmcFirmwareUpdatesParams(selected(args[0]))

			res, err := apiclient.API.GETV1BmcFirmwareUpdates(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := firmwareCheckTable.Table()
			for _, result := range res.Results {
				state := colorState(result.Success)

				updates := result.Data.Value.Updates.Value
				if len(updates) == 0 {
					t.AppendRow(result.Node, state, result.Message, "", "", "", "")
					continue
				}

				for _, update := range updates {
					t.AppendRow(
						result.Node,
						state,
						result.Message,
						update.DisplayName,
						update.InstalledVersion,
						colorVersion(update.InstalledVersion, update.PackageVersion),
						update.RebootType,
					)
				}
			}
			t.Print()

			return summaryLine(summary(res.Summary))
		},
	}
)

func init() {
	firmwareCheckTable = shared.RegisterTableFlags(firmwareCheckCmd, firmwareCheckColumns)

	firmwareCmd.AddCommand(firmwareCheckCmd)
}
