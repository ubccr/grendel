// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendelctl/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	imageCmd = &cobra.Command{
		Use:               "image {nodeset | all} <image>",
		Short:             "Change nodes boot image",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: nodesetCompletion,
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			req := &client.NodeBootImageRequest{
				Image: client.NewOptString(args[1]),
			}
			params := client.PATCHV1NodesImageParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := apiclient.API.PATCHV1NodesImage(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(res)
		},
	}
)

func init() {
	nodeCmd.AddCommand(imageCmd)
}
