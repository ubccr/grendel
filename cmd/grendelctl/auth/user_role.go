// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendelctl/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	userRoleCmd = &cobra.Command{
		Use:   "role <username> <role>",
		Short: "Edit a users role",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			req := &client.UserRoleRequest{
				Role: client.NewOptString(args[1]),
			}
			params := client.PATCHV1UsersUsernamesRoleParams{
				Usernames: args[0],
			}
			res, err := apiclient.API.PATCHV1UsersUsernamesRole(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(res)
		},
	}
)

func init() {
	userCmd.AddCommand(userRoleCmd)
}
