// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	listCmd = &cobra.Command{
		Use:               "list {nodeset | all]",
		Short:             "list nodes",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: nodesetCompletion,
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			req := client.GETV1NodesFindParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := gc.GETV1NodesFind(context.Background(), req)
			if err != nil {
				return cmd.NewApiError(err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

			fmt.Fprintln(w, "Name\tProvision\tFirmware\tBoot Image\tTags")
			for _, node := range res {
				fmt.Fprintf(w, "%s\t%t\t%s\t%s\t%s\n", node.Name.Value, node.Provision.Value, node.Firmware.Value, node.BootImage.Value, strings.Join(node.Tags.Value, ","))
			}

			return w.Flush()
		},
	}
)

func init() {
	nodeCmd.AddCommand(listCmd)
}
