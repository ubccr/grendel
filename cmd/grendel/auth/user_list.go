// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	userListTable   *shared.TableFlags
	userListColumns = []shared.Column{
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
			res, err := apiclient.API.GETV1Users(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := userListTable.Table()

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
	userListTable = shared.RegisterTableFlags(userListCmd, userListColumns)

	userCmd.AddCommand(userListCmd)
}
