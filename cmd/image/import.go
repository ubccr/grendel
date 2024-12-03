// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"context"
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/model"
)

var (
	importCmd = &cobra.Command{
		Use:   "import",
		Short: "import images",
		Long:  `import images`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			for _, name := range args {
				file, err := os.Open(name)
				if err != nil {
				}
				defer file.Close()

				cmd.Log.Infof("Processing file: %s", name)

				var images model.BootImageList
				if err := json.NewDecoder(file).Decode(&images); err != nil {
					return err
				}

				_, err = gc.ImageApi.StoreImages(context.Background(), images)
				if err != nil {
					return cmd.NewApiError("Failed to store images", err)
				}

				cmd.Log.Infof("Successfully imported %d images from: %s", len(images), name)
			}
			return nil

		},
	}
)

func init() {
	imageCmd.AddCommand(importCmd)
}
