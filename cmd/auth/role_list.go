// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	roleListFilter  []string
	roleListOptions *cmd.TableOptions
	roleListColumns = []cmd.Column{
		{Name: "Name"},
		{Name: "Permissions"},
	}
	roleListCmd = &cobra.Command{
		Use:   "list [name]...",
		Short: "List roles",
		Args:  cobra.ArbitraryArgs,
		RunE: func(command *cobra.Command, args []string) error {
			params := client.GETV1RolesParams{
				Name: client.NewOptString(strings.Join(args, ",")),
			}
			res, err := cmd.API.GETV1Roles(command.Context(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			t := cmd.NewTable(roleListColumns, roleListFilter).Options(*roleListOptions).Empty("-")

			for _, role := range res.Roles {
				permissions := make([]string, 0, len(role.PermissionList))
				for _, p := range role.PermissionList {
					permissions = append(permissions, fmt.Sprintf("%s=%s", p.Method.Value, p.Path.Value))
				}
				t.AppendRow(
					role.Name.Value,
					strings.Join(permissions, "\n"),
				)
			}

			t.Print()
			return nil
		},
	}
)

func init() {
	roleListOptions = cmd.RegisterTableFlags(roleListCmd)
	roleListCmd.Flags().StringSliceVar(&roleListFilter, "filter", nil, "filter table columns by name")
	roleListCmd.RegisterFlagCompletionFunc("filter", cmd.ColumnCompletion(roleListColumns))

	roleCmd.AddCommand(roleListCmd)
}
