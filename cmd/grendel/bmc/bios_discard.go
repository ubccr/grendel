// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var biosDiscardCmd = ackCommand(&cobra.Command{
	Use:   "discard <nodeset | all>",
	Short: "Discard staged BIOS changes",
	Args:  cobra.ExactArgs(1),
}, func(command *cobra.Command, args []string) (*client.BmcResponseModelBmcAck, error) {
	return apiclient.API.POSTV1BmcBiosDiscard(command.Context(), client.POSTV1BmcBiosDiscardParams(selected(args[0])))
})

func init() {
	biosCmd.AddCommand(biosDiscardCmd)
}
