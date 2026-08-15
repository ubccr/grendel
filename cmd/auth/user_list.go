// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	userListFilter  []string
	userListOptions *cmd.TableOptions
	userListColumns = []cmd.Column{
		{Name: "Username"},
		{Name: "Role"},
		{Name: "Enabled"},
		{Name: "Modified"},
		{Name: "Created"},
	}
	userListCmd = &cobra.Command{
		Use:   "list [username]...",
		Short: "list user(s)",
		Args:  cobra.ArbitraryArgs,
		RunE: func(command *cobra.Command, args []string) error {
			params := client.GETV1UsersParams{
				Usernames: client.NewOptString(strings.Join(args, ",")),
			}
			res, err := cmd.API.GETV1Users(command.Context(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			t := cmd.NewTable(userListColumns, userListFilter).Options(*userListOptions).Empty("-")

			for _, u := range res {
				t.AppendRow(
					u.Username.Value,
					u.Role.Value,
					fmt.Sprintf("%t", u.Enabled.Value),
					u.ModifiedAt.Value.Local().Format(time.RFC822),
					u.CreatedAt.Value.Local().Format(time.RFC822),
				)
			}

			t.Print()
			return nil
		},
	}
)

func init() {
	userListOptions = cmd.RegisterTableFlags(userListCmd)
	userListCmd.Flags().StringSliceVar(&userListFilter, "filter", nil, "filter table columns by name")
	userListCmd.RegisterFlagCompletionFunc("filter", cmd.ColumnCompletion(userListColumns))

	userCmd.AddCommand(userListCmd)
}
