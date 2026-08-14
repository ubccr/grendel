// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	powerBmcCmd = &cobra.Command{
		Use:   "reboot-bmc {nodeset | all}",
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
			res, err := cmd.API.POSTV1BmcPowerBmc(command.Context(), params)
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
	bmcCmd.AddCommand(powerBmcCmd)
}
