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
	powerFilter  []string
	powerOptions *shared.TableOptions
	powerColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Message"},
	}
	override string
	powerCmd = &cobra.Command{
		Use:   "power {cycle | off | on | redfish.ResetType} {nodeset | all}",
		Short: "Change power state of nodes",
		Long:  "Valid redfish.ResetType options: On, ForceOn, ForceOff, ForceRestart, GracefulRestart, GracefulShutdown, PowerCycle, Nmi, PushPowerButton, Suspend, Pause, Resume, FullPowerCycle",
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

			req := &client.BmcPowerBody{
				Target: client.NewOptString("host"),
				Action: client.NewOptString(powerOption),
			}
			// the server rejects an unknown boot source, so only send one when it was asked for
			if override != "" && override != "None" {
				req.BootOverride = client.NewOptString(override)
			}

			nodeset, all := selector(args[1])
			params := client.POSTV1BmcPowerParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcPower(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(powerColumns, powerFilter).Options(*powerOptions).Empty("-")
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
	powerOptions = shared.RegisterTableFlags(powerCmd)
	powerCmd.Flags().StringSliceVar(&powerFilter, "filter", nil, "filter table columns by name")
	powerCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(powerColumns))

	powerCmd.Flags().StringVarP(&override, "override", "o", "None", "Set redfish boot override. Valid options: None, Pxe, BiosSetup, Utilities, Diags")
	bmcCmd.AddCommand(powerCmd)
}
