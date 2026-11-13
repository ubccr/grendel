// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	secureBootResetFilter    []string
	secureBootResetDefault   bool
	secureBootResetDeleteAll bool
	secureBootResetDeletePK  bool
	secureBootResetYes       bool
	secureBootResetOptions   *shared.TableOptions
	secureBootResetCmd       = &cobra.Command{
		Use:   "reset <nodeset | all> [PK | KEK | db | dbx] (--to-default | --delete-all | --delete-pk)",
		Short: "Reset the secure boot databases",
		Long: `Reset the secure boot databases of every database, or of the one named.

--to-default restores the factory keys, which undoes any custom enrollment. It is the way back for a node that no longer boots because of its keys.
--delete-all removes every key, which puts the node in setup mode, ready for a new PK.
--delete-pk removes only the platform key, which also puts the node in setup mode. It applies to every database, so it takes no database.`,
		Example: `  grendel bmc secureboot reset cpn-a01-01 --to-default
  grendel bmc secureboot reset cpn-a01-01 db --delete-all`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(command *cobra.Command, args []string) error {
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
					return errors.New("--delete-pk applies to every database, leave out the database")
				}
				body.Database = client.NewOptString(args[1])
				scope = " in " + args[1]
			}

			if ok, err := confirmed(secureBootResetYes, label+scope+" on "+args[0]); !ok {
				return err
			}

			nodeset, all := selector(args[0])
			params := client.POSTV1BmcSecurebootResetParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcSecurebootReset(command.Context(), body, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}
			return printAcks(res, secureBootResetFilter, secureBootResetOptions)
		},
	}
)

func init() {
	secureBootResetOptions = shared.RegisterTableFlags(secureBootResetCmd)
	secureBootResetCmd.Flags().StringSliceVar(&secureBootResetFilter, "filter", nil, "filter table columns by name")
	secureBootResetCmd.Flags().BoolVar(&secureBootResetDefault, "to-default", false, "restore the factory keys")
	secureBootResetCmd.Flags().BoolVar(&secureBootResetDeleteAll, "delete-all", false, "delete every key, which leaves the node in setup mode")
	secureBootResetCmd.Flags().BoolVar(&secureBootResetDeletePK, "delete-pk", false, "delete only the platform key, which leaves the node in setup mode")
	secureBootResetCmd.Flags().BoolVarP(&secureBootResetYes, "yes", "y", false, "don't ask for confirmation")
	secureBootResetCmd.MarkFlagsMutuallyExclusive("to-default", "delete-all", "delete-pk")
	secureBootResetCmd.MarkFlagsOneRequired("to-default", "delete-all", "delete-pk")
	secureBootResetCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(secureBootAckColumns))
	secureBootResetCmd.ValidArgsFunction = completeDatabases(1, secureBootWritable)

	secureBootCmd.AddCommand(secureBootResetCmd)
}
