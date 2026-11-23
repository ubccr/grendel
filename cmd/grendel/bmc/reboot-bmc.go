// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var powerBmcCmd = ackCommand(&cobra.Command{
	Use:   "reboot-bmc <nodeset | all>",
	Short: "Reboot the BMC",
	Args:  cobra.ExactArgs(1),
}, func(command *cobra.Command, args []string) (*client.BmcResponseModelBmcAck, error) {
	// one power route covers both, a bmc target is always a graceful manager restart
	req := &client.BmcPowerBody{Target: client.NewOptString("bmc")}

	return apiclient.API.POSTV1BmcPower(command.Context(), req, client.POSTV1BmcPowerParams(selected(args[0])))
})

func init() {
	bmcCmd.AddCommand(powerBmcCmd)
}
