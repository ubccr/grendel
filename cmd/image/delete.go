// Copyright 2021 Grendel Authors. All rights reserved.
// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
)

var (
	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete images",
		Long:  `Delete images`,
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			_, err = gc.ImageApi.ImageDelete(context.Background(), args[0])
			if err != nil {
				return cmd.NewApiError("Failed to delete hosts", err)
			}

			fmt.Println("Successfully deleted image")

			return nil

		},
	}
)

func init() {
	imageCmd.AddCommand(deleteCmd)
}
