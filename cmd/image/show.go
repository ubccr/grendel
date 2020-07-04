// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package image

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/model"
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
