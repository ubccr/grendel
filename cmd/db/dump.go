// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package db

import (
	"context"
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump database",
		Long:  `Dump database`,
		Args:  cobra.ExactArgs(0),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			params := client.GETV1DbDumpParams{}
			res, err := gc.GETV1DbDump(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "    ")
			if err := enc.Encode(res); err != nil {
				return err
			}

			return nil

		},
	}
)

func init() {
	dbCmd.AddCommand(dumpCmd)
}
