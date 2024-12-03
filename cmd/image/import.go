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
