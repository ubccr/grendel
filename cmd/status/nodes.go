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
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/api"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

type StatTag struct {
	provision   *nodeset.NodeSet
	unprovision *nodeset.NodeSet
}

var (
	nodeLong bool
	nodesCmd = &cobra.Command{
		Use:   "nodes",
		Short: "Detailed node status",
		Long:  `Detailed node status`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			defaultImage := viper.GetString("provision.default_image")
			inputTags := strings.Join(args, ",")

			var hostList model.HostList

			if inputTags == "" {
				hostList, _, err = gc.HostApi.HostList(context.Background())
			} else {
				hostList, _, err = gc.HostApi.HostTags(context.Background(), inputTags)
			}

			if err != nil {
				return cmd.NewApiError("Failed to list hosts", err)
			}

			stats := make(map[string]*StatTag)

			nodes := 0
			for _, host := range hostList {
				for _, tag := range host.Tags {
					if inputTags != "" && !strings.Contains(inputTags, tag) {
						continue
					}

					if _, ok := stats[tag]; !ok {
						stats[tag] = &StatTag{provision: nodeset.EmptyNodeSet(), unprovision: nodeset.EmptyNodeSet()}
					}

					if host.Provision {
						stats[tag].provision.Add(host.Name)
					} else {
						stats[tag].unprovision.Add(host.Name)
					}
				}

				if len(host.Tags) == 0 {
					if _, ok := stats[""]; !ok {
						stats[""] = &StatTag{provision: nodeset.EmptyNodeSet(), unprovision: nodeset.EmptyNodeSet()}
					}

					if host.Provision {
						stats[""].provision.Add(host.Name)
					} else {
						stats[""].unprovision.Add(host.Name)
					}
				}

				nodes++
			}

			fmt.Printf("Grendel version %s\n\n", api.Version)
			fmt.Printf("Nodes: %s\n\n", humanize.Comma(int64(nodes)))

			if nodeLong {
				fmt.Printf("%-20s%-19s%-17s%-11s%-20s%-25s\n", "Name", "MAC", "IP", "Provision", "Image", "Tags")
				for _, host := range hostList {
					ipAddr := ""
					macAddr := ""
					bootNic := host.BootInterface()
					if bootNic != nil {
						ipAddr = bootNic.IP.String()
						macAddr = bootNic.MAC.String()
					}

					bi := host.BootImage

					if bi == "" {
						bi = defaultImage
					}

					printer := cyan
					if host.Provision {
						printer = yellow
					}

					printer.Printf("%-20s%-19s%-17s%-11s%-20s%-25s\n",
						host.Name,
						macAddr,
						ipAddr,
						fmt.Sprintf("%#v", host.Provision),
						bi,
						strings.Join(host.Tags, ","))

				}
				return nil
			}

			for tag, stat := range stats {
				if tag == "" {
					red.Printf("Tag: (none)\n")
				} else {
					cyan.Printf("Tag: %s\n", tag)
				}
				fmt.Printf("Provision: %s\n", stat.provision.String())
				fmt.Printf("Unprovision: %s\n", stat.unprovision.String())
				fmt.Println()
			}

			return nil
		},
	}
)

func init() {
	nodesCmd.Flags().BoolVar(&nodeLong, "long", false, "Display long format")
	statusCmd.AddCommand(nodesCmd)
}
