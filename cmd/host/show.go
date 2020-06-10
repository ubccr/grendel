// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package host

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/model"
)

var (
	showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show hosts",
		Long:  `Show hosts`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
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
