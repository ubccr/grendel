// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	userEnableCmd = &cobra.Command{
		Use:   "enabled <username> <true | false>",
		Short: "Enable or disable a user",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			b, err := strconv.ParseBool(args[1])
			if err != nil {
				return err
			}

			req := &client.UserEnableRequest{
				Enabled: client.NewOptBool(b),
			}
			params := client.PATCHV1UsersUsernamesEnableParams{
				Usernames: args[0],
			}
			res, err := cmd.API.PATCHV1UsersUsernamesEnable(command.Context(), req, params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(res)
		},
	}
)

func init() {
	userCmd.AddCommand(userEnableCmd)
}
