// Copyright 2021 Grendel Authors. All rights reserved.
// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	deleteCmd = &cobra.Command{
		Use:               "delete <nodeset>",
		Short:             "Delete nodes",
		Long:              `Delete nodes`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: nodesetCompletion,
		RunE: func(command *cobra.Command, args []string) error {
			params := client.DELETEV1NodesParams{
				Nodeset: client.NewOptString(args[0]),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := cmd.API.DELETEV1Nodes(command.Context(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(res)
		},
	}
)

func init() {
	nodeCmd.AddCommand(deleteCmd)
}
