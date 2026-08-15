// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	listFilter  []string
	listOptions *cmd.TableOptions
	listColumns = []cmd.Column{
		{Name: "Name"},
		{Name: "Provision"},
		{Name: "Firmware"},
		{Name: "Boot Image"},
		{Name: "IP"},
		{Name: "FQDN"},
		{Name: "MAC"},
		{Name: "Tags", MinWidth: 10},
	}

	listCmd = &cobra.Command{
		Use:               "list [nodeset]...",
		Short:             "list nodes",
		Aliases:           []string{"ls"},
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: nodesetCompletion,
		RunE: func(command *cobra.Command, args []string) error {
			params := client.GETV1NodesFindParams{
				Nodeset: client.NewOptString(strings.Join(args, ",")),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := cmd.API.GETV1NodesFind(command.Context(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			t := cmd.NewTable(listColumns, listFilter).Options(*listOptions).Empty("-")

			for _, node := range res {
				ips := make([]string, 0, len(node.Interfaces))
				fqdns := make([]string, 0, len(node.Interfaces))
				macs := make([]string, 0, len(node.Interfaces))
				for _, iface := range node.Interfaces {
					ips = append(ips, iface.Value.IP.Value)
					fqdns = append(fqdns, iface.Value.Fqdn.Value)
					macs = append(macs, iface.Value.MAC.Value)
				}

				t.AppendRow(
					node.Name.Value,
					strconv.FormatBool(node.Provision.Value),
					node.Firmware.Value,
					node.BootImage.Value,
					strings.Join(ips, ","),
					strings.Join(fqdns, ","),
					strings.Join(macs, ","),
					strings.Join(node.Tags.Value, ","),
				)
			}

			t.Print()

			return nil
		},
	}
)

func init() {
	listOptions = cmd.RegisterTableFlags(listCmd)
	listCmd.Flags().StringSliceVar(&listFilter, "filter", []string{"IP", "FQDN", "MAC"}, "filter table columns by name")

	listCmd.RegisterFlagCompletionFunc("filter", cmd.ColumnCompletion(listColumns))

	nodeCmd.AddCommand(listCmd)
}
