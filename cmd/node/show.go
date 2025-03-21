// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	showCmd = &cobra.Command{
		Use:   "show {nodeset | all]",
		Short: "Show nodes",
		Args:  cobra.ExactArgs(1),
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

			return output(res)
		},
	}
)

func init() {
	nodeCmd.AddCommand(showCmd)
}

func output(data any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")

	return enc.Encode(data)
}
