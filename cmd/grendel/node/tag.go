// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	tagCmd = &cobra.Command{
		Use:   "tag {nodeset | all} <tags>...",
		Short: "Tag nodes",
		Args:  cobra.MinimumNArgs(2),
		ValidArgsFunction: func(command *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return nodesetCompletion(command, args, toComplete)
			case 1:
				return tagCompletion(command, args, toComplete)
			default:
				return nil, cobra.ShellCompDirectiveNoFileComp

			}
		},
		RunE: func(command *cobra.Command, args []string) error {
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
			res, err := apiclient.API.PATCHV1NodesTagsAction(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(res)
		},
	}
)

func init() {
	nodeCmd.AddCommand(tagCmd)
}
