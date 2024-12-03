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
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/internal/bmc"
)

var (
	configureCmd = &cobra.Command{
		Use:   "configure",
		Short: "Set iDRAC to Auto configure",
		Long:  `Set iDRAC to Auto configure`,
		RunE: func(command *cobra.Command, args []string) error {
			return runConf()
		},
	}
)

func init() {
	bmcCmd.AddCommand(configureCmd)
}

func runConf() error {
	job := bmc.NewJob()

	output, err := job.BmcAutoConfigure(hostList)
	if err != nil {
		return err
	}

	bmc.PrintStatusCli(output)

	return nil
}
