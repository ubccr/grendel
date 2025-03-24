// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package status

import (
	"context"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/api"
	"github.com/ubccr/grendel/pkg/client"
	"github.com/ubccr/grendel/pkg/nodeset"
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
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			defaultImage := viper.GetString("provision.default_image")
			inputTags := strings.Join(args, ",")

			req := client.GETV1NodesFindParams{
				Nodeset: client.NewOptString(strings.Join(nodes, ",")),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			hostList, err := gc.GETV1NodesFind(context.Background(), req)
			if err != nil {
				return cmd.NewApiError(err)
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

					if host.Provision.Value {
						stats[tag].provision.Add(host.Name.Value)
					} else {
						stats[tag].unprovision.Add(host.Name.Value)
					}
				}

				if len(host.Tags) == 0 {
					if _, ok := stats[""]; !ok {
						stats[""] = &StatTag{provision: nodeset.EmptyNodeSet(), unprovision: nodeset.EmptyNodeSet()}
					}

					if host.Provision.Value {
						stats[""].provision.Add(host.Name.Value)
					} else {
						stats[""].unprovision.Add(host.Name.Value)
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
					if len(host.Interfaces) > 0 {
						ipAddr = host.Interfaces[0].Value.IP.Value
						macAddr = host.Interfaces[0].Value.MAC.Value
					}

					bi := host.BootImage.Value

					if bi == "" {
						bi = defaultImage
					}

					printer := cyan
					if host.Provision.Value {
						printer = yellow
					}

					printer.Printf("%-20s%-19s%-17s%-11s%-20s%-25s\n",
						host.Name.Value,
						macAddr,
						ipAddr,
						fmt.Sprintf("%#v", host.Provision.Value),
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
