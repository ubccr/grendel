// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	secureBootResetDefault   bool
	secureBootResetDeleteAll bool
	secureBootResetDeletePK  bool
	secureBootResetYes       bool
	secureBootResetCmd       = ackCommand(&cobra.Command{
		Use:   "reset <nodeset | all> [PK | KEK | db | dbx] (--to-default | --delete-all | --delete-pk)",
		Short: "Reset the secure boot databases",
		Long: `Reset the secure boot databases of every database, or of the one named.

--to-default restores the factory keys, which undoes any custom enrollment. It is the way back for a node that no longer boots because of its keys.
--delete-all removes every key, which puts the node in setup mode, ready for a new PK.
--delete-pk removes only the platform key, which also puts the node in setup mode. It applies to every database, so it takes no database.`,
		Args: cobra.RangeArgs(1, 2),
	}, func(command *cobra.Command, args []string) (*client.BmcResponseModelBmcAck, error) {
		var (
			typ   client.BmcSecureBootResetBodyType
			label string
		)
		switch {
		case secureBootResetDefault:
			typ, label = client.BmcSecureBootResetBodyTypeToDefault, "Restore the factory secure boot keys"
		case secureBootResetDeleteAll:
			typ, label = client.BmcSecureBootResetBodyTypeDeleteAll, "Delete every secure boot key"
		case secureBootResetDeletePK:
			typ, label = client.BmcSecureBootResetBodyTypeDeletePk, "Delete the secure boot platform key"
		}

		body := &client.BmcSecureBootResetBody{Type: client.NewOptBmcSecureBootResetBodyType(typ)}
		scope := ""
		if len(args) == 2 {
			if secureBootResetDeletePK {
				return nil, errors.New("--delete-pk applies to every database, leave out the database")
			}
			body.Database = client.NewOptString(args[1])
			scope = " in " + args[1]
		}

		if ok, err := confirmed(secureBootResetYes, label+scope+" on "+target(args[0])); !ok {
			return nil, err
		}

		return apiclient.API.POSTV1BmcSecurebootReset(command.Context(), body, client.POSTV1BmcSecurebootResetParams(selected(args[0])))
	})
)

func init() {
	secureBootResetCmd.Flags().BoolVar(&secureBootResetDefault, "to-default", false, "restore the factory keys")
	secureBootResetCmd.Flags().BoolVar(&secureBootResetDeleteAll, "delete-all", false, "delete every key, which leaves the node in setup mode")
	secureBootResetCmd.Flags().BoolVar(&secureBootResetDeletePK, "delete-pk", false, "delete only the platform key, which leaves the node in setup mode")
	secureBootResetCmd.Flags().BoolVarP(&secureBootResetYes, "yes", "y", false, "don't ask for confirmation")
	secureBootResetCmd.MarkFlagsMutuallyExclusive("to-default", "delete-all", "delete-pk")
	secureBootResetCmd.MarkFlagsOneRequired("to-default", "delete-all", "delete-pk")
	secureBootResetCmd.ValidArgsFunction = completeDatabases(1, secureBootWritable)

	secureBootCmd.AddCommand(secureBootResetCmd)
}
