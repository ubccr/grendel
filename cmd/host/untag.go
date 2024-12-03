// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package host

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
)

var (
	untagCmd = &cobra.Command{
		Use:   "untag",
		Short: "Untag hosts",
		Long:  `Untag hosts`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 || len(tags) == 0 {
				return fmt.Errorf("Please provide tags (--tags) and a nodeset")
			}

			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			nodes := strings.Join(args, ",")
			_, err = gc.HostApi.HostUntag(context.Background(), nodes, strings.Join(tags, ","))
			if err != nil {
				return cmd.NewApiError("Failed to untag hosts", err)
			}

			cmd.Log.Info("Successfully untagged hosts")

			return nil

		},
	}
)

func init() {
	hostCmd.AddCommand(untagCmd)
}
