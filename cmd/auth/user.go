// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	userCmd = &cobra.Command{
		Use:   "user",
		Short: "User commands",
	}
	userShowCmd = &cobra.Command{
		Use:   "show {username | all}",
		Short: "show user(s)",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			filter := args[0]

			params := client.GETV1UsersParams{}
			if filter != "all" {
				params.Usernames = client.NewOptString(filter)
			}
			res, err := gc.GETV1Users(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

			fmt.Fprintln(w, "Username\tRole\tEnabled\tModified\tCreated\t")
			for _, u := range res {
				fmt.Fprintf(w, "%s\t%s\t%t\t%s\t%s\t\n", u.Username.Value, u.Role.Value, u.Enabled.Value, u.ModifiedAt.Value.Local().Format(time.RFC822), u.CreatedAt.Value.Local().Format(time.RFC822))
			}

			return w.Flush()
		},
	}
	userRoleCmd = &cobra.Command{
		Use:   "role <username> <role>",
		Short: "Edit a users role",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			req := &client.UserRoleRequest{
				Role: client.NewOptString(args[1]),
			}
			params := client.PATCHV1UsersUsernamesRoleParams{
				Usernames: args[0],
			}
			res, err := gc.PATCHV1UsersUsernamesRole(context.Background(), req, params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(res)
		},
	}
	userEnableCmd = &cobra.Command{
		Use:   "enabled <username> {true | false}",
		Short: "Enable or disable a user",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

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
			res, err := gc.PATCHV1UsersUsernamesEnable(context.Background(), req, params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(res)
		},
	}
	userAddCmd = &cobra.Command{
		Use:   "add <username>",
		Short: "Add a new user",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			m := initialUserAddModel()

			pr := tea.NewProgram(m)
			om, err := pr.Run()
			if err != nil {
				return err
			}

			if m, ok := om.(userAddModel); ok && m.save {
				if len(m.inputs) != 2 {
					return errors.New("incorrect input length")
				}

				password := m.inputs[0].Value()
				if password != m.inputs[1].Value() {
					return errors.New("passwords do not match")
				}
				req := &client.AuthSignupRequest{
					Password: password,
					Username: args[0],
				}
				params := client.POSTV1AuthSignupParams{}
				res, err := gc.POSTV1AuthSignup(context.Background(), req, params)
				if err != nil {
					return cmd.NewApiError(err)
				}

				fmt.Printf("Successfully created user %s with role %s\n", res.Username.Value, res.Role.Value)
				return nil
			}

			return nil
		},
	}
	userDeleteCmd = &cobra.Command{
		Use:   "delete <username>...",
		Short: "Delete user(s)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			params := client.DELETEV1UsersUsernamesParams{
				Usernames: strings.Join(args, ","),
			}
			res, err := gc.DELETEV1UsersUsernames(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(res)
		},
	}
)

func init() {
	authCmd.AddCommand(userCmd)
	userCmd.AddCommand(userShowCmd)
	userCmd.AddCommand(userRoleCmd)
	userCmd.AddCommand(userEnableCmd)
	userCmd.AddCommand(userAddCmd)
	userCmd.AddCommand(userDeleteCmd)
}
