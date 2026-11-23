// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	override string
	powerCmd = ackCommand(&cobra.Command{
		Use:   "power {cycle | off | on | redfish.ResetType} {nodeset | all}",
		Short: "Change power state of nodes",
		Long:  "Valid redfish.ResetType options: On, ForceOn, ForceOff, ForceRestart, GracefulRestart, GracefulShutdown, PowerCycle, Nmi, PushPowerButton, Suspend, Pause, Resume, FullPowerCycle",
		Args:  cobra.ExactArgs(2),
	}, func(command *cobra.Command, args []string) (*client.BmcResponseModelBmcAck, error) {
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

		return apiclient.API.POSTV1BmcPower(command.Context(), req, client.POSTV1BmcPowerParams(selected(args[1])))
	})
)

func init() {
	powerCmd.Flags().StringVarP(&override, "override", "o", "None", "Set redfish boot override. Valid options: None, Pxe, BiosSetup, Utilities, Diags")
	bmcCmd.AddCommand(powerCmd)
}
