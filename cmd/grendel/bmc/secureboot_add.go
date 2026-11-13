// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	secureBootAddFilter   []string
	secureBootAddOwner    string
	secureBootAddHash     string
	secureBootAddHashType string
	secureBootAddOptions  *shared.TableOptions
	secureBootAddCmd      = &cobra.Command{
		Use:   "add <nodeset | all> <PK | KEK | db | dbx> [certificate file]...",
		Short: "Enroll certificates or a hash in a secure boot database",
		Long: `Enroll certificates or a hash in a secure boot database. A file can be PEM, holding one or more certificates, or a single DER certificate. A key already in the database is left alone.

Enrolling usually means signing your own kernel or shim with a key whose certificate goes in db. A node in user or deployed mode may refuse changes to PK and KEK, and on a Dell the BIOS SecureBootPolicy may need to be Custom first.

Use --hash to add a hash instead, usually to revoke a binary in dbx.`,
		Example: `  grendel bmc secureboot add cpn-a01-[01-10] db grendel-db.pem
  grendel bmc secureboot add all dbx --hash 80b4d96931bf0d02fd91a61e19d14f1da452e66db2408ca8604d411f92659f0a`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			files := args[2:]
			if secureBootAddHash != "" && len(files) > 0 {
				return errors.New("give certificate files or --hash, not both")
			}
			if secureBootAddHash == "" && len(files) == 0 {
				return errors.New("give at least one certificate file, or --hash")
			}

			body := &client.BmcSecureBootKeyBody{Database: client.NewOptString(args[1])}
			if secureBootAddOwner != "" {
				body.Owner = client.NewOptString(secureBootAddOwner)
			}
			if secureBootAddHash != "" {
				body.Signature = client.NewOptString(secureBootAddHash)
				body.SignatureType = client.NewOptBmcSecureBootKeyBodySignatureType(client.BmcSecureBootKeyBodySignatureType(secureBootAddHashType))
			}
			for _, file := range files {
				certs, err := readCertificates(file)
				if err != nil {
					return err
				}
				for _, cert := range certs {
					body.Certificates = append(body.Certificates, certPEM(cert))
					fmt.Printf("%s %s\n", certFingerprint(cert), cert.Subject.CommonName)
				}
			}
			if len(body.Certificates) > 0 {
				fmt.Println()
			}

			nodeset, all := selector(args[0])
			params := client.POSTV1BmcSecurebootKeysParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			// one request for every certificate, so a bundle is one run and one event, and each node's message says what it enrolled and skipped
			res, err := apiclient.API.POSTV1BmcSecurebootKeys(command.Context(), body, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}
			return printAcks(res, secureBootAddFilter, secureBootAddOptions)
		},
	}
)

func init() {
	secureBootAddOptions = shared.RegisterTableFlags(secureBootAddCmd)
	secureBootAddCmd.Flags().StringSliceVar(&secureBootAddFilter, "filter", nil, "filter table columns by name")
	secureBootAddCmd.Flags().StringVar(&secureBootAddOwner, "owner", "", "UEFI signature owner GUID. Defaults to whatever the BMC picks")
	secureBootAddCmd.Flags().StringVar(&secureBootAddHash, "hash", "", "hex hash to add instead of a certificate, such as the SHA-256 of a binary to revoke in dbx")
	secureBootAddCmd.Flags().StringVar(&secureBootAddHashType, "hash-type", "EFI_CERT_SHA256_GUID", "UEFI type of --hash")
	secureBootAddCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(secureBootAckColumns))
	secureBootAddCmd.RegisterFlagCompletionFunc("hash-type", cobra.FixedCompletions([]string{"EFI_CERT_SHA256_GUID", "EFI_CERT_SHA384_GUID", "EFI_CERT_SHA512_GUID", "EFI_CERT_X509_SHA256_GUID"}, cobra.ShellCompDirectiveNoFileComp))
	secureBootAddCmd.ValidArgsFunction = completeDatabases(1, secureBootWritable)

	secureBootCmd.AddCommand(secureBootAddCmd)
}
