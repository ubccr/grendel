// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
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
			if strings.ToLower(args[0]) == "all" {
				params := client.GETV1ImagesParams{}
				res, err := cmd.API.GETV1Images(command.Context(), params)
				if err != nil {
					return cmd.NewApiError(err)
				}
				return output(res)
			} else {
				params := client.GETV1ImagesFindParams{
					Names: client.NewOptString(strings.Join(args, ",")),
				}
				res, err := cmd.API.GETV1ImagesFind(command.Context(), params)
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
