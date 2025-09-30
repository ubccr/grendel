// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	editCmd = &cobra.Command{
		Use:   "edit <name>...",
		Short: "edit images",
		Long: `edit images
WARNING: Do not edit the "id" field in the JSON`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			params := client.GETV1ImagesFindParams{
				Names: client.NewOptString(strings.Join(args, ",")),
			}
			imageList, err := gc.GETV1ImagesFind(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			data, err := json.MarshalIndent(imageList, "", "    ")
			if err != nil {
				return err
			}

			newData, err := util.CaptureInputFromEditor(data)
			if err != nil {
				return err
			}

			var check []client.NilBootImageAddRequestBootImagesItem
			err = json.Unmarshal(newData, &check)
			if err != nil {
				return fmt.Errorf("invalid json. not saving changes: %w", err)
			}

			storeReq := &client.BootImageAddRequest{
				BootImages: check,
			}
			storeParams := client.POSTV1ImagesParams{}
			storeRes, err := gc.POSTV1Images(context.Background(), storeReq, storeParams)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(storeRes)
		},
	}
)

func init() {
	imageCmd.AddCommand(editCmd)
}
