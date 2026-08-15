// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package db

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendelctl/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump database",
		Args:  cobra.ExactArgs(0),
		RunE: func(command *cobra.Command, args []string) error {
			params := client.GETV1DbDumpParams{}
			res, err := apiclient.API.GETV1DbDump(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return shared.OutputJSON(res)
		},
	}
)

func init() {
	dbCmd.AddCommand(dumpCmd)
}
