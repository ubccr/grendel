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

package bmc

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/internal/bmc"
)

var (
	statusLong bool
	statusCmd  = &cobra.Command{
		Use:   "status",
		Short: "Check BMC status",
		Long:  `Check BMC status`,
		RunE: func(command *cobra.Command, args []string) error {
			return cmdStatus()
		},
	}
)

func init() {
	statusCmd.Flags().BoolVar(&statusLong, "long", false, "Display long format")
	bmcCmd.AddCommand(statusCmd)
}

func cmdStatus() error {
	job := bmc.NewJob()
	output, err := job.BmcStatus(hostList)
	if err != nil {
		return err
	}

	if statusLong {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")

		err := enc.Encode(output)
		if err != nil {
			return err
		}
	} else {
		for _, o := range output {

			if !statusLong {
				fmt.Printf("%s\t%s\t%s\n", o.Name, o.PowerStatus, o.BIOSVersion)
				continue
			}
		}

	}

	return nil
}
