// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
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
	statusOEM  bool
	statusCmd  = &cobra.Command{
		Use:   "status {nodeset | all}",
		Short: "Check BMC status",
		Long:  `Check BMC status`,
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			params := client.GETV1BmcParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := cmd.API.GETV1Bmc(command.Context(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			output := make([]client.RedfishSystem, len(res))
			for i, v := range res {
				if !statusOEM {
					v.OemDell.Reset()
				}
				output[i] = v
			}

			if statusLong {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "    ")

				err := enc.Encode(output)
				if err != nil {
					return err
				}
			} else {
				for _, o := range output {

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
	statusCmd.Flags().BoolVar(&statusOEM, "oem", false, "Display oem info")
	bmcCmd.AddCommand(statusCmd)
}
