// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package host

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/model"
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
