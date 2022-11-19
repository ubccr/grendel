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

package host

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/model"
)

var (
	bmc      bool
	tokenCmd = &cobra.Command{
		Use:   "token",
		Short: "Generate boot token for hosts",
		Long:  `Generate boot token for hosts`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			hostList, _, err := gc.HostApi.HostFind(context.Background(), strings.Join(args, ","))
			if err != nil {
				return cmd.NewApiError("Failed to find hosts for boot token generation", err)
			}

			if len(hostList) == 0 {
				return errors.New("No hosts found")
			}

			for _, host := range hostList {
				if len(host.Interfaces) == 0 {
					continue
				}

				nic := host.BootInterface()
				if bmc {
					nic = host.InterfaceBMC()
				}

				if nic == nil {
					nic = host.Interfaces[0]
				}

				token, err := model.NewBootToken(host.ID.String(), nic.MAC.String())
				if err != nil {
					return fmt.Errorf("Failed to generate signed boot token for host %s: %s", host.Name, err)
				}

				fmt.Printf("%s: %s\n", host.Name, token)
			}

			return nil

		},
	}
)

func init() {
	tokenCmd.Flags().BoolVar(&bmc, "bmc", false, "Generate token for BMC interface")
	hostCmd.AddCommand(tokenCmd)
}
