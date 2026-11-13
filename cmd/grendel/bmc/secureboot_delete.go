// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	secureBootDeleteFilter  []string
	secureBootDeleteID      string
	secureBootDeleteYes     bool
	secureBootDeleteOptions *shared.TableOptions
	secureBootDeleteCmd     = &cobra.Command{
		Use:   "delete <nodeset | all> <PK | KEK | db | dbx> [fingerprint | certificate file]",
		Short: "Remove a key from a secure boot database",
		Long: `Remove a certificate or hash from a secure boot database. Name it by its full fingerprint, as 'bmc secureboot keys --full' shows it, or by the certificate file it was enrolled from. A node that does not have the key is left alone.

--id removes a key by its id on the BMC instead. Ids differ from node to node, so use it on one node at a time.`,
		Example: `  grendel bmc secureboot delete cpn-a01-[01-10] db grendel-db.pem
  grendel bmc secureboot delete cpn-a01-01 db --id 3`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(command *cobra.Command, args []string) error {
			body := &client.BmcSecureBootDeleteBody{Database: client.NewOptString(args[1])}
			switch {
			case len(args) == 3 && secureBootDeleteID != "":
				return errors.New("give a fingerprint, a certificate file or --id, only one")
			case len(args) == 3:
				fp, err := deleteFingerprint(args[2])
				if err != nil {
					return err
				}
				body.Fingerprint = client.NewOptString(fp)
			case secureBootDeleteID != "":
				body.ID = client.NewOptString(secureBootDeleteID)
			default:
				return errors.New("give the fingerprint or certificate file of the key to remove, or --id")
			}

			if ok, err := confirmed(secureBootDeleteYes, fmt.Sprintf("Remove this key from %s on %s", args[1], args[0])); !ok {
				return err
			}

			nodeset, all := selector(args[0])
			params := client.POSTV1BmcSecurebootKeysDeleteParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcSecurebootKeysDelete(command.Context(), body, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}
			return printAcks(res, secureBootDeleteFilter, secureBootDeleteOptions)
		},
	}
)

// deleteFingerprint reads a certificate file for its fingerprint, or takes the argument as a fingerprint. Hex never holds a dot or a slash, so an argument with one is a file, and a missing one is reported as such rather than as bad hex
func deleteFingerprint(arg string) (string, error) {
	if _, err := os.Stat(arg); err != nil {
		if strings.ContainsAny(arg, "./") {
			return "", fmt.Errorf("no certificate file %s", arg)
		}
		return arg, nil
	}
	certs, err := readCertificates(arg)
	if err != nil {
		return "", err
	}
	if len(certs) > 1 {
		return "", fmt.Errorf("%s holds %d certificates, remove them one at a time by fingerprint", arg, len(certs))
	}
	return certFingerprint(certs[0]), nil
}

func init() {
	secureBootDeleteOptions = shared.RegisterTableFlags(secureBootDeleteCmd)
	secureBootDeleteCmd.Flags().StringSliceVar(&secureBootDeleteFilter, "filter", nil, "filter table columns by name")
	secureBootDeleteCmd.Flags().StringVar(&secureBootDeleteID, "id", "", "remove the key with this id on the BMC, instead of by fingerprint")
	secureBootDeleteCmd.Flags().BoolVarP(&secureBootDeleteYes, "yes", "y", false, "don't ask for confirmation")
	secureBootDeleteCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(secureBootAckColumns))
	secureBootDeleteCmd.ValidArgsFunction = completeDatabases(1, secureBootWritable)

	secureBootCmd.AddCommand(secureBootDeleteCmd)
}
