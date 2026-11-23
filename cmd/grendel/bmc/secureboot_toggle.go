// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

// secureBootToggleCmd builds enable and disable, which differ only in the value they send. Both ask first: enabling leaves a node unable to boot anything not signed by a key in db, and disabling drops that protection
func secureBootToggleCmd(enable bool) *cobra.Command {
	var yes bool

	use, short, label := "disable", "Turn UEFI secure boot off from the next boot", "Disable secure boot on "
	if enable {
		use, short, label = "enable", "Turn UEFI secure boot on from the next boot", "Enable secure boot on "
	}

	cmd := ackCommand(&cobra.Command{
		Use:   use + " <nodeset | all>",
		Short: short,
		Long:  short + ". A node only boots with secure boot on if what it boots is signed by a key in db, so check the keys first with 'bmc secureboot keys'",
		Args:  cobra.ExactArgs(1),
	}, func(command *cobra.Command, args []string) (*client.BmcResponseModelBmcAck, error) {
		if ok, err := confirmed(yes, label+target(args[0])); !ok {
			return nil, err
		}

		return apiclient.API.POSTV1BmcSecureboot(command.Context(), &client.BmcSecureBootBody{Enabled: enable}, client.POSTV1BmcSecurebootParams(selected(args[0])))
	})

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "don't ask for confirmation")
	return cmd
}

func init() {
	secureBootCmd.AddCommand(secureBootToggleCmd(true), secureBootToggleCmd(false))
}
