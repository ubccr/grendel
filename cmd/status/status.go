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

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/api"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/logger"
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
		Args:  cobra.MinimumNArgs(0),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			defaultImage := viper.GetString("provision.default_image")

			imageList, _, err := gc.ImageApi.ImageList(context.Background())
			if err != nil {
				return cmd.NewApiError("Failed to list images", err)
			}

			stats := &Stats{images: make(map[string]*StatProvision), tags: make(map[string]*StatProvision)}

			for _, img := range imageList {
				stats.images[img.Name] = &StatProvision{}
			}

			hostList, _, err := gc.HostApi.HostList(context.Background())
			if err != nil {
				return cmd.NewApiError("Failed to list hosts", err)
			}

			nodes := 0
			for _, host := range hostList {
				bi := host.BootImage

				if bi == "" {
					bi = defaultImage
				}

				if _, ok := stats.images[bi]; !ok {
					stats.images[bi] = &StatProvision{}
				}

				if host.Provision {
					stats.images[bi].provision++
				} else {
					stats.images[bi].unprovision++
				}

				for _, tag := range host.Tags {
					if _, ok := stats.tags[tag]; !ok {
						stats.tags[tag] = &StatProvision{}
					}

					if host.Provision {
						stats.tags[tag].provision++
					} else {
						stats.tags[tag].unprovision++
					}
				}

				nodes++
			}

			fmt.Printf("Grendel version %s\n\n", api.Version)
			yellow.Printf("Nodes: %s\n\n", humanize.Comma(int64(nodes)))
			fmt.Printf("%-30s%15s%15s%15s\n", fmt.Sprintf("Boot Images (%d)", len(imageList)), "Provision", "Unprovision", "Total")
			for img, stat := range stats.images {
				cyan.Printf("%-30s%15s%15s%15s\n",
					img,
					humanize.Comma(int64(stat.provision)),
					humanize.Comma(int64(stat.unprovision)),
					humanize.Comma(int64(stat.provision+stat.unprovision)))
			}

			fmt.Println()

			fmt.Printf("%-30s%15s%15s%15s\n", fmt.Sprintf("Tags (%d)", len(stats.tags)), "Provision", "Unprovision", "Total")
			for tag, stat := range stats.tags {
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
	cmd.Root.AddCommand(statusCmd)
}
