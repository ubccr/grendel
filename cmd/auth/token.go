// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	tokemCmd = &cobra.Command{
		Use:   "token <username> <role> <expire>",
		Short: "Create an auth token",
		Long: `Args:
	username:	String, must be a valid user.
	role: 		Type of model.Role, valid options: disabled, user, admin.
	expire: 	String parsed by time.ParseDuration, examples include: infinite, 8h, 30m, 20s.
		`,
		Args: cobra.MinimumNArgs(3),
		RunE: func(command *cobra.Command, args []string) error {
			req := &client.AuthTokenRequest{
				Username: client.NewOptString(args[0]),
				Role:     client.NewOptString(args[1]),
				Expire:   client.NewOptString(args[2]),
			}
			params := client.POSTV1AuthTokenParams{}
			res, err := cmd.API.POSTV1AuthToken(command.Context(), req, params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			fmt.Println(res.Token.Value)

			return nil
		},
	}
)

func init() {
	authCmd.AddCommand(tokemCmd)
}
