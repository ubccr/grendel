// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	bmc      bool
	tokenCmd = &cobra.Command{
		Use:   "token {nodeset | all} {boot | bmc}",
		Short: "Generate boot token for nodes",
		Args:  cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			params := client.GETV1NodesTokenInterfaceParams{
				Interface: args[0],
				Nodeset:   client.NewOptString(nodeset),
				Tags:      client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := gc.GETV1NodesTokenInterface(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			for _, node := range res.Nodes {
				fmt.Printf("%s: %s\n", node.Name.Value, node.Token.Value)
			}

			return nil
		},
	}
)

func init() {
	nodeCmd.AddCommand(tokenCmd)
}
