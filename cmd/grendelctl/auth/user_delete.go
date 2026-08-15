// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendelctl/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	userDeleteCmd = &cobra.Command{
		Use:   "delete <username>...",
		Short: "Delete user(s)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			params := client.DELETEV1UsersUsernamesParams{
				Usernames: strings.Join(args, ","),
			}
			res, err := apiclient.API.DELETEV1UsersUsernames(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(res)
		},
	}
)

func init() {
	userCmd.AddCommand(userDeleteCmd)
}
