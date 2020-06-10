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
	"context"
	"errors"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check BMC status",
		Long:  `Check BMC status`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return runStatus(strings.Join(args, ","))
		},
	}
)

func init() {
	bmcCmd.AddCommand(statusCmd)
}

func runStatus(ns string) error {
	gc, err := cmd.NewClient()
	if err != nil {
		return err
	}

	hostList, _, err := gc.HostApi.HostFind(context.Background(), ns)
	if err != nil {
		return cmd.NewApiError("Failed to find hosts for bmc status", err)
	}

	if len(hostList) == 0 {
		return errors.New("No hosts found")
	}

	delay := viper.GetInt("bmc.delay")
	runner := NewJobRunner(viper.GetInt("bmc.fanout"))
	for _, host := range hostList {
		runner.RunStatus(host)
		time.Sleep(time.Duration(delay) * time.Second)
	}

	runner.Wait()

	return nil
}
