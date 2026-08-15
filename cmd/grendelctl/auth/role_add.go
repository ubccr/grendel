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
	roleAddCmd = &cobra.Command{
		Use:   "add <name> [inherit]",
		Short: "Create a new role",
		Long: `Args:
	<name> is the name of the role to add
	[inherit] is the optional name of an existing role, it will set the permissions of the new role equal to the existing role`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(command *cobra.Command, args []string) error {
			req := client.PostRolesRequest{
				Role: client.NewOptString(args[0]),
			}
			if len(args) == 2 {
				req.InheritedRole = client.NewOptString(args[1])
			}
			params := client.POSTV1RolesParams{}
			res, err := apiclient.API.POSTV1Roles(command.Context(), &req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(res)
		},
	}
)

func init() {
	roleCmd.AddCommand(roleAddCmd)
}
