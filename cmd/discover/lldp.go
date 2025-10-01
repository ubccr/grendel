// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package discover

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	lldpPorts string
	lldpLong  bool
	lldpCmd   = &cobra.Command{
		Use:   "lldp <switch name>",
		Short: "Query LLDP port info from a switch",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			p := client.GETV1SwitchNodesetLldpParams{
				Nodeset: args[0],
				Ports:   client.NewOptString(lldpPorts),
			}
			res, err := gc.GETV1SwitchNodesetLldp(context.Background(), p)
			if err != nil {
				return cmd.NewApiError(err)
			}
			header := "Port\tMAC Address\tSystem Name"
			headerLong := "Port\tMAC Address\tSystem Name\tSystem Port ID\tManagement Address\tSystem Port Description\tSystem Description"

			if lldpLong {
				header = headerLong
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
			fmt.Fprintln(w, header)
			for _, lldp := range res {
				if lldpLong {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", lldp.PortName.Value, lldp.ChassisID.Value, lldp.SystemName.Value, lldp.PortID.Value, lldp.ManagementAddress.Value, lldp.PortDescription.Value, lldp.SystemDescription.Value)
					continue
				}

				fmt.Fprintf(w, "%s\t%s\t%s\n", lldp.PortName.Value, lldp.ChassisID.Value, lldp.SystemName.Value)
			}
			return w.Flush()
		},
	}
)

func init() {
	lldpCmd.PersistentFlags().StringVarP(&lldpPorts, "ports", "p", "", "Comma separated list of ports. Should match the switch port naming convention")
	lldpCmd.PersistentFlags().BoolVar(&lldpLong, "long", false, "Display full LLDP info")
	discoverCmd.AddCommand(lldpCmd)
}
