// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	listTable   *shared.TableFlags
	listColumns = []shared.Column{
		{Name: "Name"},
		{Name: "Provision"},
		{Name: "Firmware"},
		{Name: "Arch"},
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
			res, err := apiclient.API.GETV1NodesFind(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := listTable.Table()

			for _, node := range res {
				ips := make([]string, 0, len(node.Interfaces))
				fqdns := make([]string, 0, len(node.Interfaces))
				macs := make([]string, 0, len(node.Interfaces))
				for _, iface := range node.Interfaces {
					ips = append(ips, iface.IP.Value)
					fqdns = append(fqdns, iface.Fqdn.Value)
					macs = append(macs, iface.MAC.Value)
				}

				t.AppendRow(
					node.Name.Value,
					strconv.FormatBool(node.Provision.Value),
					node.Firmware.Value,
					node.Arch.Value,
					node.BootImage.Value,
					strings.Join(ips, "\n"),
					strings.Join(fqdns, "\n"),
					strings.Join(macs, "\n"),
					strings.Join(node.Tags.Value, ","),
				)
			}

			t.Print()

			return nil
		},
	}
)

func init() {
	listTable = shared.RegisterTableFlags(listCmd, listColumns, "IP", "FQDN", "MAC")

	nodeCmd.AddCommand(listCmd)
}
