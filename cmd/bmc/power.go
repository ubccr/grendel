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
	powerCmd = &cobra.Command{
		Use:   "power",
		Short: "Change power state of hosts",
		Long:  `Change power state of hosts`,
	}
	cycleCmd = &cobra.Command{
		Use:   "cycle",
		Short: "Reboot the hosts",
		Long:  `Reboot the hosts`,
		RunE: func(command *cobra.Command, args []string) error {
			return runPowerCycle()
		},
	}
	onCmd = &cobra.Command{
		Use:   "on",
		Short: "Power on the hosts",
		Long:  `Power on the hosts`,
		RunE: func(command *cobra.Command, args []string) error {
			return runPowerOn()
		},
	}
	offCmd = &cobra.Command{
		Use:   "off",
		Short: "Power off the hosts",
		Long:  `Power off the hosts`,
		RunE: func(command *cobra.Command, args []string) error {
			return runPowerOff()
		},
	}

	override string
)

func init() {
	bmcCmd.AddCommand(powerCmd)
	powerCmd.AddCommand(cycleCmd)
	powerCmd.AddCommand(onCmd)
	powerCmd.AddCommand(offCmd)
	cycleCmd.Flags().StringVar(&override, "override", "none", "Set Boot override option. Valid options: none, pxe, bios-setup hdd, usb, diagnostics, utilities")
	onCmd.Flags().StringVar(&override, "override", "none", "Set Boot override option. Valid options: none, pxe, bios-setup hdd, usb, diagnostics, utilities")
}

func runPowerCycle() error {
	job := bmc.NewJob()

	output, err := job.PowerCycle(hostList, override)
	if err != nil {
		return err
	}

	bmc.PrintStatusCli(output)

	return nil
}

func runPowerOn() error {
	job := bmc.NewJob()

	output, err := job.PowerOn(hostList, override)
	if err != nil {
		return err
	}

	bmc.PrintStatusCli(output)

	return nil
}

func runPowerOff() error {
	job := bmc.NewJob()

	output, err := job.PowerOff(hostList)
	if err != nil {
		return err
	}

	bmc.PrintStatusCli(output)

	return nil
}
