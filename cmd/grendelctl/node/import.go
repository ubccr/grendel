// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendelctl/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	importCmd = &cobra.Command{
		Use:   "import <filenames>...",
		Short: "import nodes",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			var nodes []client.NilNodeAddRequestNodeListItem
			for _, name := range args {
				file, err := os.Open(name)
				if err != nil {
					fmt.Printf("failed to open file. name=%s err=%s", name, err)
				}
				defer file.Close()

				var node []client.NilNodeAddRequestNodeListItem
				if err := json.NewDecoder(file).Decode(&node); err != nil {
					return err
				}

				nodes = append(nodes, node...)
			}
			req := &client.NodeAddRequest{
				NodeList: nodes,
			}
			params := client.POSTV1NodesParams{}
			res, err := apiclient.API.POSTV1Nodes(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(res)
		},
	}
)

func init() {
	nodeCmd.AddCommand(importCmd)
}
