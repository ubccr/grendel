// Copyright 2021 Grendel Authors. All rights reserved.
// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	deleteCmd = &cobra.Command{
		Use:   "delete <name>...",
		Short: "Delete images",
		Long:  `Delete images`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			params := client.DELETEV1ImagesParams{
				Names: client.NewOptString(strings.Join(args, ",")),
			}
			res, err := apiclient.API.DELETEV1Images(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(res)
		},
	}
)

func init() {
	imageCmd.AddCommand(deleteCmd)
}
