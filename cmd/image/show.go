// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/model"
)

var (
	showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show images",
		Long:  `Show images`,
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			var imageList model.BootImageList

			if strings.ToLower(args[0]) == "all" {
				imageList, _, err = gc.ImageApi.ImageList(context.Background())
				if err != nil {
					return cmd.NewApiError("Failed to list images", err)
				}
			} else {
				imageList, _, err = gc.ImageApi.ImageFind(context.Background(), args[0])
				if err != nil {
					return cmd.NewApiError("Failed to find images", err)
				}
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "    ")
			if err := enc.Encode(imageList); err != nil {
				return err
			}

			return nil

		},
	}
)

func init() {
	imageCmd.AddCommand(showCmd)
}
