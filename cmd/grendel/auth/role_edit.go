// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	nonInteractive bool
	roleEditCmd    = &cobra.Command{
		Use:   "edit <role> [<add | remove> <method=path>...]",
		Short: "Adds or removes permissions from a role",
		Long:  "use edit <role> to interactively edit role permissions, or use edit <role> <add | remove> <method=path> -n for non-interactive usage",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			params := client.GETV1RolesParams{
				Name: client.NewOptString(args[0]),
			}
			res, err := apiclient.API.GETV1Roles(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			if len(res.Roles) != 1 {
				return errors.New("failed to find role")
			}

			if nonInteractive {
				if len(args) < 3 {
					return errors.New("invalid syntax")
				}
				role := args[0]
				var permissions []client.PatchRolesRequestPermissionListItem
				for _, p := range res.Roles[0].PermissionList {
					permissions = append(permissions, client.PatchRolesRequestPermissionListItem(p))
				}
				for _, s := range args[2:] {
					kvSlice := strings.Split(s, "=")
					if len(kvSlice) != 2 {
						return errors.New("failed to parse key value pairs")
					}
					kvPerms := client.PatchRolesRequestPermissionListItem{
						Method: client.NewOptString(kvSlice[0]),
						Path:   client.NewOptString(kvSlice[1]),
					}
					switch args[1] {
					case "add":
						permissions = append(permissions, kvPerms)
					case "remove":
						permissions = slices.DeleteFunc(permissions, func(rp client.PatchRolesRequestPermissionListItem) bool {
							return (rp.Method == kvPerms.Method && rp.Path == kvPerms.Path)
						})
					default:
						return fmt.Errorf("supported operations are 'add' or 'remove' got: %s", args[1])
					}
				}

				req := client.PatchRolesRequest{
					Role:           client.NewOptString(role),
					PermissionList: permissions,
				}
				params := client.PATCHV1RolesParams{}
				patchRes, err := apiclient.API.PATCHV1Roles(command.Context(), &req, params)
				if err != nil {
					return apiclient.NewApiError(err)
				}

				return apiclient.NewApiResponse(patchRes)
			}

			m := InitialModel(args[0])
			for i, p := range res.Roles[0].PermissionList {
				m.choices = append(m.choices, fmt.Sprintf("%s=%s", p.Method.Value, p.Path.Value))
				m.selected[i] = struct{}{}
			}

			for _, p := range res.Roles[0].UnassignedPermissionList {
				m.choices = append(m.choices, fmt.Sprintf("%s=%s", p.Method.Value, p.Path.Value))
			}

			m.paginator.SetTotalPages(len(m.choices))

			pr := tea.NewProgram(m)
			om, err := pr.Run()
			if err != nil {
				return err
			}

			if m, ok := om.(model); ok && m.save {
				var permissions []client.PatchRolesRequestPermissionListItem

				for i := range m.selected {
					cArr := strings.Split(m.choices[i], "=")
					if len(cArr) != 2 {
						return fmt.Errorf("failed to split choice string: %s", m.choices[i])
					}
					permissions = append(permissions, client.PatchRolesRequestPermissionListItem{
						Method: client.NewOptString(cArr[0]),
						Path:   client.NewOptString(cArr[1]),
					})
				}

				req := client.PatchRolesRequest{
					Role:           client.NewOptString(m.role),
					PermissionList: permissions,
				}
				params := client.PATCHV1RolesParams{}
				res, err := apiclient.API.PATCHV1Roles(command.Context(), &req, params)
				if err != nil {
					return apiclient.NewApiError(err)
				}

				return apiclient.NewApiResponse(res)
			}

			return nil
		},
	}
)

func init() {
	roleCmd.AddCommand(roleEditCmd)
	roleEditCmd.PersistentFlags().BoolVarP(&nonInteractive, "non-interactive", "n", false, "allow setting permissions with key value pairs")
}
