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
	provisionCmd = &cobra.Command{
		Use:   "provision",
		Short: "Set hosts to provision",
		Long:  `Set hosts to provision`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 && len(tags) == 0 {
				return fmt.Errorf("Please provide tags (--tags) or a nodeset")
			}

			if len(args) > 0 && len(tags) > 0 {
				log.Warn("Using both tags (--tags) and a nodeset is not supported yet. Only nodeset is used.")
			}

			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			nodes := strings.Join(args, ",")
			if len(tags) > 0 && len(args) == 0 {
				hostList, _, err := gc.HostApi.HostTags(context.Background(), strings.Join(tags, ","))
				if err != nil {
					return cmd.NewApiError("Failed to find hosts by tag", err)
				}

				ns, err := hostList.ToNodeSet()
				if err != nil {
					return cmd.NewApiError("Failed to create nodeset from host list", err)
				}

				nodes = ns.String()
			}

			_, err = gc.HostApi.HostProvision(context.Background(), nodes)
			if err != nil {
				return cmd.NewApiError("Failed to set hosts to provision", err)
			}

			cmd.Log.Info("Successfully set hosts to provision")

			return nil

		},
	}
)

func init() {
	hostCmd.AddCommand(provisionCmd)
}
