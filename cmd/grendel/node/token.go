// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	bmc       bool
	tokenArch string
	tokenCmd  = &cobra.Command{
		Use:   "token {nodeset | all} {boot | bmc}",
		Short: "Generate boot token for nodes",
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: func(command *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return nodesetCompletion(command, args, toComplete)
			case 1:
				return []string{"boot", "bmc"}, cobra.ShellCompDirectiveNoFileComp
			default:
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
		},
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			params := client.GETV1NodesTokenInterfaceParams{
				Interface: args[1],
				Nodeset:   client.NewOptString(nodeset),
				Tags:      client.NewOptString(strings.Join(tags, ",")),
				Arch:      client.NewOptString(tokenArch),
			}
			res, err := apiclient.API.GETV1NodesTokenInterface(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			for _, node := range res.Nodes {
				fmt.Printf("%s: %s\n", node.Name.Value, node.Token.Value)
			}

			return nil
		},
	}
)

func init() {
	tokenCmd.Flags().StringVar(&tokenArch, "arch", "", "Arch to embed in the token, simulating the arch detected during a PXE boot. Nodes with an arch set manually ignore it. Example: x86_64, aarch64")

	nodeCmd.AddCommand(tokenCmd)
}
