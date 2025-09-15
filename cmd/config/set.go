// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	setCmd = &cobra.Command{
		Use:   "set <key>=<value>...",
		Short: "Update configuration value",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			cfg := make(client.ConfigSetRequestUpdateConfig, 0)
			for _, a := range args {
				s := strings.Split(a, "=")
				if len(s) != 2 {
					return fmt.Errorf("invalid key value pair: %s", a)
				}
				cfg[s[0]] = s[1]
			}

			params := client.PATCHV1ConfigSetParams{}
			req := client.ConfigSetRequest{
				UpdateConfig: client.NewOptConfigSetRequestUpdateConfig(cfg),
			}

			res, err := gc.PATCHV1ConfigSet(context.Background(), &req, params)
			if err != nil {
				return err
			}

			return cmd.NewApiResponse(res)
		},
	}
)

func init() {
	cfgCmd.AddCommand(setCmd)
}
