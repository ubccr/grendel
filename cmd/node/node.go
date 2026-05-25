// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	tags    []string
	log     = logger.GetLogger("NODE")
	nodeCmd = &cobra.Command{
		Use:   "node",
		Short: "Node commands",
		Long:  `Node commands`,
	}
)

func init() {
	nodeCmd.PersistentFlags().StringSliceVarP(&tags, "tags", "t", []string{}, "filter by tags")
	cmd.Root.AddCommand(nodeCmd)
}

func nodesetCompletion(command *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	gc, err := cmd.NewOgenClient()
	if err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	tags, err := command.Flags().GetStringSlice("tags")
	if err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	req := client.GETV1NodesFindParams{
		Tags: client.NewOptString(strings.Join(tags, ",")),
	}
	res, err := gc.GETV1NodesFind(context.Background(), req)
	if err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	nodeNames := make([]string, len(res))
	for i, node := range res {
		nodeNames[i] = node.Name.Value
	}

	return nodeNames, cobra.ShellCompDirectiveNoFileComp
}

func tagCompletion(command *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	gc, err := cmd.NewOgenClient()
	if err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	req := client.GETV1NodesFindParams{
		Nodeset: client.NewOptString(args[0]),
	}
	res, err := gc.GETV1NodesFind(context.Background(), req)
	if err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	tags := make([]string, 0)

	for _, node := range res {
		tags = append(tags, node.Tags.Value...)
	}

	return tags, cobra.ShellCompDirectiveNoFileComp
}
