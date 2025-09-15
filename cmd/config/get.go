// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	getCmd = &cobra.Command{
		Use:   "get <key>",
		Short: "Get configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			params := client.GETV1ConfigGetParams{
				Key: client.NewOptString(args[0]),
			}

			res, err := gc.GETV1ConfigGet(context.Background(), params)
			if err != nil {
				return err
			}
			for _, v := range res.Config.Value {
				fmt.Println(v)
			}
			return nil
		},
	}
)

func init() {
	cfgCmd.AddCommand(getCmd)
}
