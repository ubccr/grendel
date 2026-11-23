// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var configureImportCmd = ackCommand(&cobra.Command{
	Use:   "import <nodeset | all> <NoReboot | Graceful | Forced> <filename>",
	Short: "Import BMC configuration",
	Long: `Args:
	shutdownType:	action bmc should take, NoReboot will wait for the node to be rebooted manually before applying. Any other type WILL REBOOT THE NODE.
	filename:	idrac config file relative to grendel template folder.
		`,
	Args: cobra.ExactArgs(3),
}, func(command *cobra.Command, args []string) (*client.BmcResponseModelBmcAck, error) {
	req := &client.BmcImportConfigBody{
		ShutdownType: client.NewOptString(args[1]),
		File:         client.NewOptString(args[2]),
	}

	return apiclient.API.POSTV1BmcConfigureImport(command.Context(), req, client.POSTV1BmcConfigureImportParams(selected(args[0])))
})

func init() {
	configureCmd.AddCommand(configureImportCmd)
}
