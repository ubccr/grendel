// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	unprovisionCmd = &cobra.Command{
		Use:   "unprovision {nodeset | all}",
		Short: "Unprovision nodes",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			req := &client.NodeProvisionRequest{
				Provision: client.NewOptBool(false),
			}
			params := client.PATCHV1NodesProvisionParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := gc.PATCHV1NodesProvision(context.Background(), req, params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(res)
		},
	}
)

func init() {
	nodeCmd.AddCommand(unprovisionCmd)
}
