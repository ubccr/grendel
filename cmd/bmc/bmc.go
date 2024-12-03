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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/model"
)

const (
	PowerOn int = iota
	PowerOff
	PowerCycle
)

var (
	hostList    model.HostList
	tags        []string
	bmcUser     string
	bmcPassword string
	bmcCmd      = &cobra.Command{
		Use:   "bmc",
		Short: "Query BMC devices",
		Long:  `Query BMC devices`,
	}
)

func init() {
	bmcCmd.PersistentFlags().String("user", "", "bmc user name")
	viper.BindPFlag("bmc.user", bmcCmd.PersistentFlags().Lookup("user"))
	bmcCmd.PersistentFlags().String("password", "", "bmc password")
	viper.BindPFlag("bmc.password", bmcCmd.PersistentFlags().Lookup("password"))
	bmcCmd.PersistentFlags().Int("delay", 0, "delay")
	viper.BindPFlag("bmc.delay", bmcCmd.PersistentFlags().Lookup("delay"))
	bmcCmd.PersistentFlags().Int("fanout", 1, "fanout")
	viper.BindPFlag("bmc.fanout", bmcCmd.PersistentFlags().Lookup("fanout"))

	bmcCmd.PersistentFlags().StringSliceVarP(&tags, "tags", "t", []string{}, "select nodes by tags")

	bmcCmd.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		err := cmd.SetupLogging()
		if err != nil {
			return err
		}

		bmcUser = viper.GetString("bmc.user")
		if bmcUser == "" {
			return errors.New("please set bmc user")
		}
		bmcPassword = viper.GetString("bmc.password")
		if bmcPassword == "" {
			return errors.New("please set bmc password")
		}

		if len(args) == 0 && len(tags) == 0 {
			return fmt.Errorf("please provide tags (--tags) or a nodeset")
		}

		if len(args) > 0 && len(tags) > 0 {
			cmd.Log.Warn("Using both tags (--tags) and a nodeset is not supported yet. Only nodeset is used.")
		}

		gc, err := cmd.NewClient()
		if err != nil {
			return err
		}

		if len(tags) > 0 && len(args) == 0 {
			hostList, _, err = gc.HostApi.HostTags(context.Background(), strings.Join(tags, ","))
			if err != nil {
				return cmd.NewApiError("Failed to find hosts by tag", err)
			}
		} else {
			nodes := strings.Join(args, ",")
			hostList, _, err = gc.HostApi.HostFind(context.Background(), nodes)
			if err != nil {
				return cmd.NewApiError("Failed to find hosts", err)
			}
		}

		if len(hostList) == 0 {
			return errors.New("no hosts found")
		}

		return nil
	}

	cmd.Root.AddCommand(bmcCmd)
}
