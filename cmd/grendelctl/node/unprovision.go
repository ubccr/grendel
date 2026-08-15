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
	unprovisionCmd = &cobra.Command{
		Use:               "unprovision {nodeset | all}",
		Short:             "Unprovision nodes",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: nodesetCompletion,
		RunE: func(command *cobra.Command, args []string) error {
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
			res, err := apiclient.API.PATCHV1NodesProvision(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(res)
		},
	}
)

func init() {
	nodeCmd.AddCommand(unprovisionCmd)
}
