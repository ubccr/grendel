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
	tagCmd = &cobra.Command{
		Use:   "tag {nodeset | all} <tags>...",
		Short: "Tag nodes",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			req := &client.NodeTagsRequest{
				Tags: client.NewOptString(strings.Join(args[1:], ",")),
			}
			params := client.PATCHV1NodesTagsActionParams{
				Action:  "add",
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := gc.PATCHV1NodesTagsAction(context.Background(), req, params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(res)
		},
	}
)

func init() {
	nodeCmd.AddCommand(tagCmd)
}
