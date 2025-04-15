// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	roleCmd = &cobra.Command{
		Use:   "role",
		Short: "Role commands",
	}
	showPermissions bool
	roleShowCmd     = &cobra.Command{
		Use:   "show {name | all}",
		Short: "List roles",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			var params client.GETV1RolesParams
			if args[0] != "all" {
				params.Name.SetTo(args[0])
			}
			res, err := gc.GETV1Roles(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}
			for _, role := range res.Roles {
				fmt.Println(role.Name.Value)
				if showPermissions {
					for _, p := range role.PermissionList {
						fmt.Printf("\t%s:%s\n", p.Method.Value, p.Path.Value)
					}
				}
			}

			return nil
		},
	}
	roleAddCmd = &cobra.Command{
		Use:   "add <name> [inherit]",
		Short: "Create a new role",
		Long: `Args:
	<name> is the name of the role to add
	[inherit] is the optional name of an existing role, it will set the permissions of the new role equal to the existing role`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			req := client.PostRolesRequest{
				Role: client.NewOptString(args[0]),
			}
			if len(args) == 2 {
				req.InheritedRole = client.NewOptString(args[1])
			}
			params := client.POSTV1RolesParams{}
			res, err := gc.POSTV1Roles(context.Background(), &req, params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(res)
		},
	}
	roleDeleteCmd = &cobra.Command{
		Use:   "delete <name>...",
		Short: "Delete a new role",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			names := strings.Join(args, ",")

			params := client.DELETEV1RolesNamesParams{
				Names: names,
			}
			res, err := gc.DELETEV1RolesNames(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(res)
		},
	}
	nonInteractive bool
	roleEditCmd    = &cobra.Command{
		Use:   "edit {<role> | {add | remove} <method=path>...}",
		Short: "Adds or removes permissions from a role",
		Long:  "use edit <role> to interactively edit role permissions, or use edit {add | remove} <method=path> -n for non-interactive usage",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			m := InitialModel(args[0])

			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			params := client.GETV1RolesParams{
				Name: client.NewOptString(args[0]),
			}
			res, err := gc.GETV1Roles(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
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
					if args[1] == "remove" {
						for i, ep := range permissions {
							if ep.Method == kvPerms.Method && ep.Path == kvPerms.Path {
								permissions = slices.Delete(permissions, i, i+1)
							}
						}
					} else {
						permissions = append(permissions, kvPerms)
					}
				}

				req := client.PatchRolesRequest{
					Role:           client.NewOptString(role),
					PermissionList: permissions,
				}
				params := client.PATCHV1RolesParams{}
				patchRes, err := gc.PATCHV1Roles(context.Background(), &req, params)
				if err != nil {
					return cmd.NewApiError(err)
				}

				return cmd.NewApiResponse(patchRes)
			}

			for i, p := range res.Roles[0].PermissionList {
				m.choices = append(m.choices, fmt.Sprintf("%s:%s", p.Method.Value, p.Path.Value))
				m.selected[i] = struct{}{}
			}

			for _, p := range res.Roles[0].UnassignedPermissionList {
				m.choices = append(m.choices, fmt.Sprintf("%s:%s", p.Method.Value, p.Path.Value))
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
					cArr := strings.Split(m.choices[i], ":")
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
				res, err := gc.PATCHV1Roles(context.Background(), &req, params)
				if err != nil {
					return cmd.NewApiError(err)
				}

				return cmd.NewApiResponse(res)
			}

			return nil
		},
	}
)

func init() {
	authCmd.AddCommand(roleCmd)
	roleCmd.AddCommand(roleShowCmd)
	roleCmd.AddCommand(roleAddCmd)
	roleCmd.AddCommand(roleDeleteCmd)
	roleCmd.AddCommand(roleEditCmd)
	roleEditCmd.PersistentFlags().BoolVarP(&nonInteractive, "non-interactive", "n", false, "allow setting permissions with key value pairs")
	roleShowCmd.PersistentFlags().BoolVarP(&showPermissions, "show-permissions", "p", false, "show a list of permissions under each role")
}
