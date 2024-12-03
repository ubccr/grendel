// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

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
