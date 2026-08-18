// Copyright 2021 Grendel Authors. All rights reserved.
// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	newFirmware  string
	newProvision bool
	newBootImage string
	newTags      []string
	newCmd       = &cobra.Command{
		Use:   "add <name>",
		Short: "add node",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			newNode := []client.NilNodeAddRequestNodeListItem{
				client.NewNilNodeAddRequestNodeListItem(client.NodeAddRequestNodeListItem{
					Name:      client.NewOptString(args[0]),
					Firmware:  client.NewOptString(newFirmware),
					Provision: client.NewOptBool(newProvision),
					Tags:      client.NewOptNilStringArray(newTags),
					BootImage: client.NewOptString(newBootImage),
				}),
			}

			storeReq := &client.NodeAddRequest{
				NodeList: newNode,
			}
			params := client.POSTV1NodesParams{}
			storeRes, err := apiclient.API.POSTV1Nodes(command.Context(), storeReq, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(storeRes)
		},
	}
)

func init() {
	newCmd.PersistentFlags().StringVar(&newFirmware, "firmware", "", "iPXE firmware override")
	newCmd.PersistentFlags().BoolVar(&newProvision, "provision", false, "Set the node to provision")
	newCmd.PersistentFlags().StringArrayVar(&newTags, "tags", []string{}, "Tags to add to the node. Can be passed multiple times")
	newCmd.PersistentFlags().StringVar(&newBootImage, "image", "", "Node boot image")

	nodeCmd.AddCommand(newCmd)
}
