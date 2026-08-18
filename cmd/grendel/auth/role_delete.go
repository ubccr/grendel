// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	roleDeleteCmd = &cobra.Command{
		Use:   "delete <name>...",
		Short: "Delete a role",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			names := strings.Join(args, ",")

			params := client.DELETEV1RolesNamesParams{
				Names: names,
			}
			res, err := apiclient.API.DELETEV1RolesNames(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(res)
		},
	}
)

func init() {
	roleCmd.AddCommand(roleDeleteCmd)
}
