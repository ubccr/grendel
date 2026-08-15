// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendelctl/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	showCmd = &cobra.Command{
		Use:               "show [nodeset]...",
		Short:             "Show nodes",
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: nodesetCompletion,
		RunE: func(command *cobra.Command, args []string) error {
			params := client.GETV1NodesFindParams{
				Nodeset: client.NewOptString(strings.Join(args, ",")),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := apiclient.API.GETV1NodesFind(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return shared.OutputJSON(res)
		},
	}
)

func init() {
	nodeCmd.AddCommand(showCmd)
}
