// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

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
