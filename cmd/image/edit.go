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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/model"
)

var (
	editCmd = &cobra.Command{
		Use:   "edit",
		Short: "edit images",
		Long:  `edit images`,
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			imageList, _, err := gc.ImageApi.ImageFind(context.Background(), args[0])
			if err != nil {
				return cmd.NewApiError("Failed to find images to edit", err)
			}

			data, err := json.MarshalIndent(imageList, "", "    ")
			if err != nil {
				return err
			}

			newData, err := util.CaptureInputFromEditor(data)
			if err != nil {
				return err
			}

			var check model.BootImageList
			err = json.Unmarshal(newData, &check)
			if err != nil {
				return fmt.Errorf("Invalid JSON. Not saving changes: %w", err)
			}

			_, err = gc.ImageApi.StoreImages(context.Background(), check)
			if err != nil {
				return cmd.NewApiError("Failed to store images", err)
			}

			fmt.Printf("Successfully saved %d images\n", len(check))

			return nil
		},
	}
)

func init() {
	imageCmd.AddCommand(editCmd)
}
