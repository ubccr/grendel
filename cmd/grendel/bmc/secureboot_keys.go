// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	secureBootKeysFilter   []string
	secureBootKeysDefaults bool
	secureBootKeysFull     bool
	secureBootKeysOptions  *shared.TableOptions
	secureBootKeysColumns  = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Database"},
		{Name: "ID"},
		{Name: "Kind"},
		{Name: "Subject"},
		{Name: "Issuer"},
		{Name: "Expires"},
		{Name: "Fingerprint"},
		{Name: "Owner"},
		{Name: "Message"},
	}
	secureBootKeysCmd = &cobra.Command{
		Use:   "keys <nodeset | all> [PK | KEK | db | dbx]...",
		Short: "List the keys in the secure boot databases",
		Long:  "List the certificates and hashes in the secure boot databases. A certificate's fingerprint is its SHA-256, which is the same on every node, so it is what 'bmc secureboot delete' takes. A hash's fingerprint is the hash itself",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset, all := selector(args[0])
			params := client.GETV1BmcSecurebootParams{
				Keys:    client.NewOptBool(true),
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			if len(args) > 1 {
				params.Databases = client.NewOptString(strings.Join(args[1:], ","))
			}
			if secureBootKeysDefaults {
				params.Defaults = client.NewOptBool(true)
			}

			res, err := apiclient.API.GETV1BmcSecureboot(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			now := time.Now()
			t := shared.NewTable(secureBootKeysColumns, secureBootKeysFilter).Options(*secureBootKeysOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				node, state := result.Node.Value, colorState(result.Success.Value)

				rows := 0
				for _, db := range result.Data.Value.Keys {
					for _, c := range db.Certificates {
						t.AppendRow(node, state, db.ID.Value, c.ID.Value, "cert", c.Subject.Value, c.Issuer.Value,
							expires(c.NotAfter.Value, now), shortFingerprint(c.Fingerprint.Value, secureBootKeysFull), c.Owner.Value, result.Message.Value)
						rows++
					}
					for _, s := range db.Signatures {
						t.AppendRow(node, state, db.ID.Value, s.ID.Value, hashKind(s.Type.Value), "", "", "",
							shortFingerprint(s.Hash.Value, secureBootKeysFull), s.Owner.Value, result.Message.Value)
						rows++
					}
				}
				// a node with no keys, or one that failed, still gets a row saying so
				if rows == 0 {
					t.AppendRow(node, state, "", "", "", "", "", "", "", "", result.Message.Value)
				}
			}
			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

// expires shows the date a certificate runs out, in red once it has. UEFI firmware does not check expiry, but an expired key is usually one due for replacing
func expires(notAfter string, now time.Time) string {
	when, err := time.Parse(time.RFC3339, notAfter)
	if err != nil {
		return notAfter
	}
	day := when.Format("2006-01-02")
	if when.Before(now) {
		return text.FgHiRed.Sprint(day)
	}
	return day
}

// shortFingerprint keeps the first 16 digits unless full is set, enough to tell keys apart in a table
func shortFingerprint(fp string, full bool) string {
	if full || len(fp) <= 16 {
		return fp
	}
	return fp[:16]
}

// hashKind names a signature by its hash, so EFI_CERT_SHA256_GUID shows as sha256
func hashKind(typ string) string {
	kind := strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(typ, "EFI_CERT_"), "_GUID"))
	if kind == "" {
		return "hash"
	}
	return kind
}

func init() {
	secureBootKeysOptions = shared.RegisterTableFlags(secureBootKeysCmd)
	secureBootKeysCmd.Flags().StringSliceVar(&secureBootKeysFilter, "filter", []string{"ID", "Owner", "Message"}, "filter table columns by name")
	secureBootKeysCmd.Flags().BoolVar(&secureBootKeysDefaults, "defaults", false, "also list the factory default databases, such as dbDefault")
	secureBootKeysCmd.Flags().BoolVar(&secureBootKeysFull, "full", false, "show whole fingerprints instead of the first 16 digits")
	secureBootKeysCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(secureBootKeysColumns))
	secureBootKeysCmd.ValidArgsFunction = func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return append(secureBootWritable, "PKDefault", "KEKDefault", "dbDefault", "dbxDefault"), cobra.ShellCompDirectiveNoFileComp
	}

	secureBootCmd.AddCommand(secureBootKeysCmd)
}
