// Copyright 2021 Grendel Authors. All rights reserved.
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
	deleteCmd = &cobra.Command{
		Use:               "delete <nodeset>...",
		Short:             "Delete nodes",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: nodesetCompletion,
		RunE: func(command *cobra.Command, args []string) error {
			params := client.DELETEV1NodesParams{
				Nodeset: client.NewOptString(strings.Join(args, ",")),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := apiclient.API.DELETEV1Nodes(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(res)
		},
	}
)

func init() {
	nodeCmd.AddCommand(deleteCmd)
}
