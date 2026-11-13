// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

// secureBootToggleCmd builds enable and disable, which differ only in the value they send
func secureBootToggleCmd(enable bool) *cobra.Command {
	var (
		filter  []string
		yes     bool
		options *shared.TableOptions
	)

	use, short := "disable", "Turn UEFI secure boot off from the next boot"
	if enable {
		use, short = "enable", "Turn UEFI secure boot on from the next boot"
	}

	cmd := &cobra.Command{
		Use:   use + " <nodeset | all>",
		Short: short,
		Long:  short + ". A node only boots with secure boot on if what it boots is signed by a key in db, so check the keys first with 'bmc secureboot keys'",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if !enable {
				if ok, err := confirmed(yes, "Turn secure boot off on "+args[0]); !ok {
					return err
				}
			}

			nodeset, all := selector(args[0])
			params := client.POSTV1BmcSecurebootParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcSecureboot(command.Context(), &client.BmcSecureBootBody{Enabled: enable}, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}
			return printAcks(res, filter, options)
		},
	}

	options = shared.RegisterTableFlags(cmd)
	cmd.Flags().StringSliceVar(&filter, "filter", nil, "filter table columns by name")
	cmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(secureBootAckColumns))
	if !enable {
		cmd.Flags().BoolVarP(&yes, "yes", "y", false, "don't ask for confirmation")
	}
	return cmd
}

func init() {
	secureBootCmd.AddCommand(secureBootToggleCmd(true), secureBootToggleCmd(false))
}
