// Copyright 2021 Grendel Authors. All rights reserved.
// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	deleteCmd = &cobra.Command{
		Use:   "delete <name>...",
		Short: "Delete images",
		Long:  `Delete images`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			params := client.DELETEV1ImagesParams{
				Names: client.NewOptString(strings.Join(args, ",")),
			}
			res, err := gc.DELETEV1Images(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(res)
		},
	}
)

func init() {
	imageCmd.AddCommand(deleteCmd)
}
