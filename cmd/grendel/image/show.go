// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	showCmd = &cobra.Command{
		Use:   "show [name]...",
		Short: "Show images",
		Args:  cobra.ArbitraryArgs,
		RunE: func(command *cobra.Command, args []string) error {
			params := client.GETV1ImagesFindParams{
				Names: client.NewOptString(strings.Join(args, ",")),
			}
			res, err := apiclient.API.GETV1ImagesFind(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return shared.OutputJSON(res)
		},
	}
)

func init() {
	imageCmd.AddCommand(showCmd)
}
