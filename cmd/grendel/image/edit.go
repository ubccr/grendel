// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	editCmd = &cobra.Command{
		Use:   "edit <name>...",
		Short: "edit images",
		Long:  `edit images`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			p := client.GETV1ImagesFindParams{
				Names: client.NewOptString(strings.Join(args, ",")),
			}
			originalImages, err := apiclient.API.GETV1ImagesFind(command.Context(), p)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			originalText, err := json.MarshalIndent(originalImages, "", "    ")
			if err != nil {
				return err
			}

			var response *client.GenericResponse
			err = util.EditLoop(originalText, func(editedText []byte) error {
				var editedJson []client.NilBootImageAddRequestBootImagesItem
				err = json.Unmarshal(editedText, &editedJson)
				if err != nil {
					return err
				}

				// Verify IDs
				for _, originalImage := range originalImages {
					if !slices.ContainsFunc(editedJson, func(editedJson client.NilBootImageAddRequestBootImagesItem) bool {
						return editedJson.Value.ID.Value == originalImage.ID.Value
					}) {
						return errors.New("failed to save, id field cannot be modified")
					}
				}

				response, err = apiclient.API.POSTV1Images(command.Context(), &client.BootImageAddRequest{
					BootImages: editedJson,
				}, client.POSTV1ImagesParams{})
				if err != nil {
					return apiclient.NewApiError(err)
				}

				return nil
			})
			if err != nil {
				return err
			}

			if response != nil {
				return apiclient.NewApiResponse(response)
			}
			return nil
		},
	}
)

func init() {
	imageCmd.AddCommand(editCmd)
}
