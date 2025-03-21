// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	importCmd = &cobra.Command{
		Use:   "import <filenames>...",
		Short: "import nodes",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			for _, name := range args {
				file, err := os.Open(name)
				if err != nil {
					fmt.Printf("failed to open file. name=%s err=%s", name, err)
				}
				defer file.Close()

				var nodes []client.NilNodeAddRequestNodeListItem
				if err := json.NewDecoder(file).Decode(&nodes); err != nil {
					return err
				}

				req := &client.NodeAddRequest{
					NodeList: nodes,
				}
				params := client.POSTV1NodesParams{}
				res, err := gc.POSTV1Nodes(context.Background(), req, params)
				if err != nil {
					return cmd.NewApiError(err)
				}

				err = cmd.NewApiResponse(res)
				if err != nil {
					return err
				}
			}
			return nil

		},
	}
)

func init() {
	nodeCmd.AddCommand(importCmd)
}
