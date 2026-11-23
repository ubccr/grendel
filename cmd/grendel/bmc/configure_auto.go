// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var configureAutoCmd = ackCommand(&cobra.Command{
	Use:   "auto <nodeset | all>",
	Short: "Set iDRAC to Auto configure",
	Args:  cobra.ExactArgs(1),
}, func(command *cobra.Command, args []string) (*client.BmcResponseModelBmcAck, error) {
	return apiclient.API.POSTV1BmcConfigureAuto(command.Context(), client.POSTV1BmcConfigureAutoParams(selected(args[0])))
})

func init() {
	configureCmd.AddCommand(configureAutoCmd)
}
