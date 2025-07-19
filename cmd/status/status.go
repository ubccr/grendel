// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package status

import (
	"context"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/api"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/pkg/client"
)

type StatProvision struct {
	provision   int
	unprovision int
}

type Stats struct {
	images map[string]*StatProvision
	tags   map[string]*StatProvision
}

var (
	tags      []string
	nodes     []string
	log       = logger.GetLogger("STATUS")
	cyan      = color.New(color.FgCyan)
	green     = color.New(color.FgGreen)
	red       = color.New(color.FgRed)
	yellow    = color.New(color.FgYellow)
	blue      = color.New(color.FgBlue)
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Status commands",
		Long:  `Status commands`,
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			defaultImage := viper.GetString("provision.default_image")
			inputTags := strings.Join(args, ",")

			params := client.GETV1ImagesParams{}
			imageList, err := gc.GETV1Images(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			stats := &Stats{images: make(map[string]*StatProvision), tags: make(map[string]*StatProvision)}

			for _, img := range imageList {
				stats.images[img.Name] = &StatProvision{}
			}

			req := client.GETV1NodesFindParams{
				Nodeset: client.NewOptString(strings.Join(nodes, ",")),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			hostList, err := gc.GETV1NodesFind(context.Background(), req)
			if err != nil {
				return cmd.NewApiError(err)
			}

			nodes := 0
			for _, host := range hostList {
				bi := host.BootImage.Value

				if bi == "" {
					bi = defaultImage
				}

				if _, ok := stats.images[bi]; !ok {
					stats.images[bi] = &StatProvision{}
				}

				if host.Provision.Value {
					stats.images[bi].provision++
				} else {
					stats.images[bi].unprovision++
				}

				for _, tag := range host.Tags.Value {
					if inputTags != "" && !strings.Contains(inputTags, tag) {
						continue
					}

					if _, ok := stats.tags[tag]; !ok {
						stats.tags[tag] = &StatProvision{}
					}

					if host.Provision.Value {
						stats.tags[tag].provision++
					} else {
						stats.tags[tag].unprovision++
					}
				}

				if len(host.Tags.Value) == 0 {
					if _, ok := stats.tags[""]; !ok {
						stats.tags[""] = &StatProvision{}
					}

					if host.Provision.Value {
						stats.tags[""].provision++
					} else {
						stats.tags[""].unprovision++
					}
				}

				nodes++
			}

			fmt.Printf("Grendel version %s\n\n", api.Version)
			yellow.Printf("Nodes: %s\n\n", humanize.Comma(int64(nodes)))

			if inputTags == "" {
				fmt.Printf("%-30s%15s%15s%15s\n", fmt.Sprintf("Boot Images (%d)", len(imageList)), "Provision", "Unprovision", "Total")
				for img, stat := range stats.images {
					cyan.Printf("%-30s%15s%15s%15s\n",
						img,
						humanize.Comma(int64(stat.provision)),
						humanize.Comma(int64(stat.unprovision)),
						humanize.Comma(int64(stat.provision+stat.unprovision)))
				}

				fmt.Println()
				return nil
			}

			fmt.Printf("%-30s%15s%15s%15s\n", fmt.Sprintf("Tags (%d)", len(stats.tags)), "Provision", "Unprovision", "Total")
			for _, tag := range strings.Split(inputTags, ",") {
				if _, ok := stats.tags[tag]; !ok {
					continue
				}

				stat := stats.tags[tag]

				cyan.Printf("%-30s%15s%15s%15s\n",
					tag,
					humanize.Comma(int64(stat.provision)),
					humanize.Comma(int64(stat.unprovision)),
					humanize.Comma(int64(stat.provision+stat.unprovision)))
			}

			return nil
		},
	}
)

func init() {
	statusCmd.PersistentFlags().StringSliceVarP(&tags, "tags", "t", []string{}, "filter by tags")
	statusCmd.PersistentFlags().StringSliceVarP(&nodes, "nodes", "n", []string{}, "filter by nodeset")
	cmd.Root.AddCommand(statusCmd)
}
