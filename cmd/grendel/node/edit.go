// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	editCmd = &cobra.Command{
		Use:               "edit [nodeset]...",
		Short:             "edit nodes",
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: nodesetCompletion,
		RunE: func(command *cobra.Command, args []string) error {
			p := client.GETV1NodesFindParams{
				Nodeset: client.NewOptString(strings.Join(args, ",")),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			originalNodes, err := apiclient.API.GETV1NodesFind(command.Context(), p)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			originalText, err := json.MarshalIndent(originalNodes, "", "    ")
			if err != nil {
				return err
			}

			var response *client.GenericResponse
			err = util.EditLoop(originalText, func(editedText []byte) error {
				var editedJson []client.NilNodeAddRequestNodeListItem
				err = json.Unmarshal(editedText, &editedJson)
				if err != nil {
					return err
				}

				// Verify IDs
				for _, originalNode := range originalNodes {
					for _, editedNode := range editedJson {
						if originalNode.Name.Value != editedNode.Value.Name.Value {
							continue
						}
						// Compare node ID
						if originalNode.ID.Value != editedNode.Value.ID.Value {
							return errors.New("failed to save, id field cannot be modified")
						}

						// Compare interface IDs
						for _, originalInterface := range originalNode.Interfaces {
							found := false
							for _, editedInterfaces := range editedNode.Value.Interfaces {
								if originalInterface.Value.ID != editedInterfaces.Value.ID {
									continue
								}
								found = true
							}
							// Only error if ID is BOTH not found and no interfaces were removed (allow iface delete)
							if !found && (len(originalNode.Interfaces) == len(editedNode.Value.Interfaces)) {
								return errors.New("failed to save, interface id fields cannot be modified")
							}
						}

						// Compare bond IDs
						for _, originalBond := range originalNode.Bonds {
							found := false
							for _, editedBonds := range editedNode.Value.Bonds {
								if originalBond.Value.ID != editedBonds.Value.ID {
									continue
								}
								found = true
							}
							if !found && (len(originalNode.Bonds) == len(editedNode.Value.Bonds)) {
								return errors.New("failed to save, bond id fields cannot be modified")
							}
						}
					}
				}

				response, err = apiclient.API.POSTV1Nodes(command.Context(), &client.NodeAddRequest{
					NodeList: editedJson,
				}, client.POSTV1NodesParams{})
				if err != nil {
					return apiclient.NewApiError(err)
				}

				return nil
			})
			if err != nil {
				return err
			}

			if response != nil {
				return apiclient.NewApiResponse(response)
			}
			return nil
		},
	}
)

func init() {
	nodeCmd.AddCommand(editCmd)
}
