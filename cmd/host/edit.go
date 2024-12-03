// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package host

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/model"
)

var (
	editCmd = &cobra.Command{
		Use:   "edit",
		Short: "edit hosts",
		Long:  `edit hosts`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) == 0 && len(tags) == 0 {
				return fmt.Errorf("Please provide tags (--tags) or a nodeset")
			}

			if len(args) > 0 && len(tags) > 0 {
				log.Warn("Using both tags (--tags) and a nodeset is not supported yet. Only nodeset is used.")
			}

			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			var hostList model.HostList

			if len(args) == 1 && strings.ToLower(args[0]) == "all" {
				hostList, _, err = gc.HostApi.HostList(context.Background())
				if err != nil {
					return cmd.NewApiError("Failed to fetch all hosts", err)
				}
			} else if len(tags) > 0 && len(args) == 0 {
				hostList, _, err = gc.HostApi.HostTags(context.Background(), strings.Join(tags, ","))
				if err != nil {
					return cmd.NewApiError("Failed to fetch hosts by tag", err)
				}
			} else {
				nodes := strings.Join(args, ",")
				hostList, _, err = gc.HostApi.HostFind(context.Background(), nodes)
				if err != nil {
					return cmd.NewApiError("Failed to fetch hosts", err)
				}
			}

			data, err := json.MarshalIndent(hostList, "", "    ")
			if err != nil {
				return err
			}

			newData, err := util.CaptureInputFromEditor(data)
			if err != nil {
				return err
			}

			var check model.HostList
			err = json.Unmarshal(newData, &check)
			if err != nil {
				return fmt.Errorf("Invalid JSON. Not saving changes: %w", err)
			}

			_, err = gc.HostApi.StoreHosts(context.Background(), check)
			if err != nil {
				return cmd.NewApiError("Failed to store hosts", err)
			}

			fmt.Printf("Successfully saved %d hosts\n", len(check))

			return nil
		},
	}
)

func init() {
	hostCmd.AddCommand(editCmd)
}
