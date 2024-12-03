// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package host

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/model"
)

var (
	showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show hosts",
		Long:  `Show hosts`,
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
					return cmd.NewApiError("Failed to list hosts", err)
				}
			} else if len(tags) > 0 && len(args) == 0 {
				hostList, _, err = gc.HostApi.HostTags(context.Background(), strings.Join(tags, ","))
				if err != nil {
					return cmd.NewApiError("Failed to find hosts by tag", err)
				}
			} else {
				nodes := strings.Join(args, ",")
				hostList, _, err = gc.HostApi.HostFind(context.Background(), nodes)
				if err != nil {
					return cmd.NewApiError("Failed to find hosts", err)
				}
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "    ")
			if err := enc.Encode(hostList); err != nil {
				return err
			}

			return nil

		},
	}
)

func init() {
	hostCmd.AddCommand(showCmd)
}
