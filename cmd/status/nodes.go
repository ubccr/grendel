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
	"github.com/ubccr/grendel/internal/api"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/model"
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
					if len(host.Interfaces) > 0 {
						ipAddr = host.Interfaces[0].CIDR()
						macAddr = host.Interfaces[0].MAC.String()
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
