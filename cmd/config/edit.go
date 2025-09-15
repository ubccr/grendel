// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	configType string
	editCmd    = &cobra.Command{
		Use:   "edit",
		Short: "edit config",
		Args:  cobra.ExactArgs(0),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			var params client.GETV1ConfigGetFileParams
			if configType != "" {
				params.Type = client.NewOptString(configType)
			}
			res, err := gc.GETV1ConfigGetFile(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			updatedConfig, err := util.CaptureInputFromEditor(res.Config)
			if err != nil {
				return err
			}

			var check []client.NilBootImageAddRequestBootImagesItem
			err = json.Unmarshal(updatedConfig, &check)
			if err != nil {
				return fmt.Errorf("Invalid JSON. Not saving changes: %w", err)
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
			// return nil
		},
	}
)

func init() {
	cfgCmd.AddCommand(editCmd)
	editCmd.PersistentFlags().StringVar(&configType, "config-type", "", "type of config file to edit. valid options: toml, json, yaml")
}
