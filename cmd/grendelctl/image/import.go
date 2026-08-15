// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendelctl/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	importCmd = &cobra.Command{
		Use:   "import <filenames>...",
		Short: "import images",
		Long:  `import images`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			var images []client.NilBootImageAddRequestBootImagesItem
			for _, name := range args {
				file, err := os.Open(name)
				if err != nil {
					fmt.Printf("failed to open file. name=%s err=%s", name, err)
				}
				defer file.Close()

				shared.Log.Infof("Processing file: %s", name)

				var image []client.NilBootImageAddRequestBootImagesItem
				if err := json.NewDecoder(file).Decode(&image); err != nil {
					return err
				}

				images = append(images, image...)
			}

			req := &client.BootImageAddRequest{
				BootImages: images,
			}
			params := client.POSTV1ImagesParams{}
			res, err := apiclient.API.POSTV1Images(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(res)

		},
	}
)

func init() {
	imageCmd.AddCommand(importCmd)
}
