// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	override string
	powerCmd = &cobra.Command{
		Use:   "power {cycle | off | on | redfish.ResetType} {nodeset | all}",
		Short: "Change power state of nodes",
		Long:  "Valid redfish.ResetType options: On, ForceOn, ForceOff, ForceRestart, GracefulRestart, GracefulShutdown, PowerCycle",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			var err error
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			// shorthand option syntax
			powerOption := ""
			switch args[0] {
			case "cycle":
				powerOption = "ForceRestart"
			case "off":
				powerOption = "ForceOff"
			case "on":
				powerOption = "ForceOn"
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
			res, err := gc.POSTV1BmcPowerOs(context.Background(), req, params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			for _, jobMessage := range res {
				fmt.Printf("%s\t %s\t %s\n", jobMessage.Host.Value, jobMessage.Status.Value, jobMessage.Msg.Value)
			}
			return nil
		},
	}
)

func init() {
	powerCmd.PersistentFlags().StringVarP(&override, "override", "o", "None", "Set redfish boot override. Valid options: None, Pxe, BiosSetup, Utilities, Diags")
	bmcCmd.AddCommand(powerCmd)
}
