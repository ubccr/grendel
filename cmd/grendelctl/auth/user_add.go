// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendelctl/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	userAddCmd = &cobra.Command{
		Use:   "add <username>",
		Short: "Add a new user",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
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
				res, err := apiclient.API.POSTV1AuthSignup(command.Context(), req, params)
				if err != nil {
					return apiclient.NewApiError(err)
				}

				fmt.Printf("Successfully created user %s with role %s\n", res.Username.Value, res.Role.Value)
				return nil
			}

			return nil
		},
	}
)

func init() {
	userCmd.AddCommand(userAddCmd)
}
