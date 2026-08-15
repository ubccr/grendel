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
	powerFilter  []string
	powerOptions *shared.TableOptions
	powerColumns = []shared.Column{
		{Name: "Node"},
		{Name: "Status"},
		{Name: "Message"},
	}
	override string
	powerCmd = &cobra.Command{
		Use:   "power {cycle | off | on | redfish.ResetType} {nodeset | all}",
		Short: "Change power state of nodes",
		Long:  "Valid redfish.ResetType options: On, ForceOn, ForceOff, ForceRestart, GracefulRestart, GracefulShutdown, PowerCycle",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			// shorthand option syntax
			powerOption := ""
			switch args[0] {
			case "cycle":
				powerOption = "ForceRestart"
			case "off":
				powerOption = "ForceOff"
			case "on":
				powerOption = "On" // really? ForceOn isn't supported Dell???
			default:
				powerOption = args[0]
			}

			nodeset := args[1]
			if args[1] == "all" {
				nodeset = ""
			}
			req := &client.BmcOsPowerBody{
				PowerOption: client.NewOptString(powerOption),
				BootOption:  client.NewOptString(override),
			}

			params := client.POSTV1BmcPowerOsParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := apiclient.API.POSTV1BmcPowerOs(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(powerColumns, powerFilter).Options(*powerOptions).Empty("-")

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
	powerOptions = shared.RegisterTableFlags(powerCmd)
	powerCmd.Flags().StringSliceVar(&powerFilter, "filter", nil, "filter table columns by name")
	powerCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(powerColumns))

	powerCmd.Flags().StringVarP(&override, "override", "o", "None", "Set redfish boot override. Valid options: None, Pxe, BiosSetup, Utilities, Diags")
	bmcCmd.AddCommand(powerCmd)
}
