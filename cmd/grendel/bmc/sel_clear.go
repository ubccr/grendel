// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var selClearCmd = ackCommand(&cobra.Command{
	Use:   "clear <nodeset | all>",
	Short: "Clear the system event log on the BMC",
	Args:  cobra.ExactArgs(1),
}, func(command *cobra.Command, args []string) (*client.BmcResponseModelBmcAck, error) {
	return apiclient.API.POSTV1BmcSelClear(command.Context(), client.POSTV1BmcSelClearParams(selected(args[0])))
})

func init() {
	selCmd.AddCommand(selClearCmd)
}
