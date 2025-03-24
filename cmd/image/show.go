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
	"github.com/ubccr/grendel/pkg/client"
)

var (
	showCmd = &cobra.Command{
		Use:   "show {names... | all}",
		Short: "Show images",
		Long:  `Show images`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			if strings.ToLower(args[0]) == "all" {
				params := client.GETV1ImagesParams{}
				res, err := gc.GETV1Images(context.Background(), params)
				if err != nil {
					return cmd.NewApiError(err)
				}
				return output(res)
			} else {
				params := client.GETV1ImagesFindParams{
					Names: client.NewOptString(strings.Join(args, ",")),
				}
				res, err := gc.GETV1ImagesFind(context.Background(), params)
				if err != nil {
					return cmd.NewApiError(err)
				}
				return output(res)
			}
		},
	}
)

func init() {
	imageCmd.AddCommand(showCmd)
}

func output(data any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")

	return enc.Encode(data)
}
