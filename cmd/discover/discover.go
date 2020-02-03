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

package discover

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/model"
)

var (
	subnetStr     string
	noProvision   bool
	subnet        net.IP
	firmwareBuild firmware.Build
	hostFile      string
	hosts         map[string]*model.Host
	log           = logger.GetLogger("DISCOVER")
	discoverCmd   = &cobra.Command{
		Use:   "discover",
		Short: "Auto-discover commands",
		Long:  `Auto-discover commands`,
	}
)

func init() {
	discoverCmd.PersistentFlags().StringP("domain", "d", "", "domain name")
	viper.BindPFlag("discovery.domain", discoverCmd.PersistentFlags().Lookup("domain"))
	discoverCmd.PersistentFlags().String("firmware", "", "firmware")
	viper.BindPFlag("discovery.firmware", discoverCmd.PersistentFlags().Lookup("firmware"))

	discoverCmd.PersistentFlags().StringVar(&hostFile, "hosts", "", "existing hosts file to add to")
	discoverCmd.PersistentFlags().BoolVar(&noProvision, "disable-provision", false, "don't set host to provision")
	discoverCmd.PersistentFlags().StringVarP(&subnetStr, "subnet", "s", "", "subnet to use for auto ip assignment (/24)")
	discoverCmd.MarkFlagRequired("subnet")

	discoverCmd.PersistentPostRunE = func(command *cobra.Command, args []string) error {
		hostList := make(model.HostList, 0)
		for _, host := range hosts {
			hostList = append(hostList, host)
		}

		if len(hostList) > 0 {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "    ")
			if err := enc.Encode(hostList); err != nil {
				return err
			}
		}

		return nil
	}

	discoverCmd.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		err := cmd.SetupLogging()
		if err != nil {
			return err
		}

		subnet = net.IPv4(0, 0, 0, 0)
		if subnetStr != "" {
			subnet = net.ParseIP(subnetStr)
			if subnet == nil || subnet.To4() == nil {
				return fmt.Errorf("Invalid IPv4 subnet address: %s", subnetStr)
			}
		}

		firmwareStr := viper.GetString("discovery.firmware")
		if firmwareStr != "" {
			firmwareBuild = firmware.NewFromString(firmwareStr)
			if firmwareBuild.IsNil() {
				return fmt.Errorf("Invalid firmware build: %s", firmwareStr)
			}
		}

		hosts = make(map[string]*model.Host)

		if hostFile != "" {
			err := loadHosts(hostFile)
			if err != nil {
				return fmt.Errorf("Invalid host file %s: %w", hostFile, err)
			}
		}

		return nil
	}

	cmd.Root.AddCommand(discoverCmd)
}

func addNic(name, fqdn string, mac net.HardwareAddr, ip net.IP, isBMC bool) {
	host, ok := hosts[name]
	if !ok {
		host = &model.Host{
			Name:       name,
			Provision:  !noProvision,
			Firmware:   firmwareBuild,
			Interfaces: make([]*model.NetInterface, 0),
		}
	}

	nic := host.Interface(mac)
	if nic == nil {
		nic = &model.NetInterface{
			MAC:  mac,
			IP:   ip,
			FQDN: fqdn,
			BMC:  isBMC,
		}

		host.Interfaces = append(host.Interfaces, nic)
	} else {
		nic.IP = ip
		nic.FQDN = fqdn
		nic.BMC = isBMC
	}

	hosts[name] = host
}

func loadHosts(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var hostList model.HostList
	err = json.Unmarshal(jsonBytes, &hostList)
	if err != nil {
		return err
	}

	for _, host := range hostList {
		hosts[host.Name] = host
	}

	return nil
}
