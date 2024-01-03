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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish/redfish"
	"github.com/ubccr/grendel/bmc"
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
			return runPower("reboot")
		},
	}
	onCmd = &cobra.Command{
		Use:   "on",
		Short: "Power on the hosts",
		Long:  `Power on the hosts`,
		RunE: func(command *cobra.Command, args []string) error {
			return runPower("on")
		},
	}
	offCmd = &cobra.Command{
		Use:   "off",
		Short: "Power off the hosts",
		Long:  `Power off the hosts`,
		RunE: func(command *cobra.Command, args []string) error {
			return runPower("off")
		},
	}

	override string
)

func init() {
	bmcCmd.AddCommand(powerCmd)
	powerCmd.AddCommand(cycleCmd)
	powerCmd.AddCommand(onCmd)
	powerCmd.AddCommand(offCmd)
	cycleCmd.Flags().StringVar(&override, "override", "none", "Set Boot override option. Valid options: none, pxe, biosSetup hdd, usb, diags, utilities")
	onCmd.Flags().StringVar(&override, "override", "none", "Set Boot override option. Valid options: none, pxe, biosSetup hdd, usb, diags, utilities")
}

func runPower(powerType string) error {
	delay := viper.GetInt("bmc.delay")
	fanout := viper.GetInt("bmc.fanout")
	runner := bmc.NewJobRunner(fanout)
	for i, host := range hostList {
		ch := make(chan string)

		boot := redfish.Boot{
			BootSourceOverrideTarget:  redfish.NoneBootSourceOverrideTarget,
			BootSourceOverrideEnabled: redfish.OnceBootSourceOverrideEnabled,
		}

		// TODO: think about how to move these switches into the runner, deduping the code
		switch override {
		case "pxe":
			boot.BootSourceOverrideTarget = redfish.PxeBootSourceOverrideTarget
		case "biosSetup":
			boot.BootSourceOverrideTarget = redfish.BiosSetupBootSourceOverrideTarget
		case "usb":
			boot.BootSourceOverrideTarget = redfish.UsbBootSourceOverrideTarget
		case "hdd":
			boot.BootSourceOverrideTarget = redfish.HddBootSourceOverrideTarget
		case "utilities":
			boot.BootSourceOverrideTarget = redfish.UtilitiesBootSourceOverrideTarget
		case "diags":
			boot.BootSourceOverrideTarget = redfish.DiagsBootSourceOverrideTarget
		}

		switch powerType {
		case "reboot":
			runner.RunPowerCycle(host, ch, boot)
		case "on":
			runner.RunPowerOn(host, ch, boot)
		case "off":
			runner.RunPowerOff(host, ch)
		}

		output := strings.Split(<-ch, "|")

		if len(output) < 3 {
			return errors.New("failed to run power job, index out of range")
		}

		fmt.Printf("%s\t%s\t%s", output[0], output[1], output[2])

		if (i+1)%fanout == 0 {
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}

	runner.Wait()

	return nil
}
