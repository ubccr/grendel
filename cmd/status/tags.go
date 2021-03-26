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

package status

import (
	"context"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/api"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/nodeset"
)

type StatTag struct {
	provision   *nodeset.NodeSet
	unprovision *nodeset.NodeSet
}

var (
	tagLong bool
	tagCmd  = &cobra.Command{
		Use:   "tags",
		Short: "Status by tag",
		Long:  `Status by tag`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			hostList, _, err := gc.HostApi.HostTags(context.Background(), strings.Join(args, ","))
			if err != nil {
				return cmd.NewApiError("Failed to list hosts by tags", err)
			}

			stats := make(map[string]*StatTag)

			nodes := 0
			for _, host := range hostList {
				for _, tag := range host.Tags {
					if _, ok := stats[tag]; !ok {
						stats[tag] = &StatTag{provision: nodeset.EmptyNodeSet(), unprovision: nodeset.EmptyNodeSet()}
					}

					if host.Provision {
						stats[tag].provision.Add(host.Name)
					} else {
						stats[tag].unprovision.Add(host.Name)
					}
				}

				nodes++
			}

			fmt.Printf("Grendel version %s\n\n", api.Version)
			yellow.Printf("Nodes: %s\n\n", humanize.Comma(int64(nodes)))

			if tagLong {
				fmt.Printf("%-30s%-19s%-17s%-10s%-25s\n", "Name", "MAC", "IP", "Provision", "Tags")
				for _, host := range hostList {
					ipAddr := ""
					macAddr := ""
					bootNic := host.BootInterface()
					if bootNic != nil {
						ipAddr = bootNic.IP.String()
						macAddr = bootNic.MAC.String()
					}
					cyan.Printf("%-30s%-19s%-17s%-11s%-25s\n",
						host.Name,
						macAddr,
						ipAddr,
						fmt.Sprintf("%#v", host.Provision),
						strings.Join(host.Tags, ","))

				}
				return nil
			}

			for tag, stat := range stats {
				cyan.Printf("Tag: %s\n", tag)
				fmt.Printf("Provision: %s\n", stat.provision.String())
				fmt.Printf("Unprovision: %s\n", stat.unprovision.String())
				fmt.Println()
			}

			return nil
		},
	}
)

func init() {
	tagCmd.Flags().BoolVar(&tagLong, "long", false, "Display long format")
	statusCmd.AddCommand(tagCmd)
}
