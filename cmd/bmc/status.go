// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	statusLong bool
	statusCmd  = &cobra.Command{
		Use:   "status {nodeset | all}",
		Short: "Check BMC status",
		Long:  `Check BMC status`,
		RunE: func(command *cobra.Command, args []string) error {
			var err error
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			params := client.GETV1BmcParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := gc.GETV1Bmc(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			if statusLong {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "    ")

				err := enc.Encode(res)
				if err != nil {
					return err
				}
			} else {
				for _, o := range res {

					if !statusLong {
						fmt.Printf("%s\t %s\t %s\t %s\n", o.Name.Value, o.PowerStatus.Value, o.SerialNumber.Value, o.BiosVersion.Value)
						continue
					}
				}

			}

			return nil
		},
	}
)

func init() {
	statusCmd.Flags().BoolVar(&statusLong, "long", false, "Display long format")
	bmcCmd.AddCommand(statusCmd)
}
