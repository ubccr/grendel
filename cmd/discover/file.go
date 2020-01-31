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
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/model"
)

var (
	fileCmd = &cobra.Command{
		Use:   "file",
		Short: "Discover hosts from file",
		Long:  `Discover hosts from file`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			hostList := make(model.HostList, 0)
			for _, name := range args {
				file, err := os.Open(name)
				if err != nil {
				}
				defer file.Close()

				cmd.Log.Infof("Processing file: %s", name)
				hosts, err := discoverFromTSV(file)
				if err != nil {
					return err
				}

				hostList = append(hostList, hosts...)

				cmd.Log.Infof("Successfully imported hosts from: %s", name)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "    ")
			if err := enc.Encode(hostList); err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	discoverCmd.AddCommand(fileCmd)
}

func discoverFromTSV(reader io.Reader) (model.HostList, error) {
	hostList := make(model.HostList, 0)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		cols := strings.Split(scanner.Text(), "\t")
		if len(cols) < 3 {
			return nil, fmt.Errorf("Invalid record format. Must be at least 3 cols name|mac|ip: %s", line)
		}

		host := &model.Host{Name: cols[0]}

		hwaddr, err := net.ParseMAC(cols[1])
		if err != nil {
			return nil, fmt.Errorf("Malformed hardware address: %s", cols[0])
		}
		ipaddr := net.ParseIP(cols[2])
		if ipaddr.To4() == nil {
			return nil, fmt.Errorf("Invalid IPv4 address: %v", cols[1])
		}

		nic := &model.NetInterface{MAC: hwaddr, IP: ipaddr}

		if len(cols) > 3 {
			nic.FQDN = cols[3]
		}

		host.Interfaces = []*model.NetInterface{nic}

		if len(cols) > 4 && strings.ToLower(cols[4]) == "yes" {
			host.Provision = true
		}

		hostList = append(hostList, host)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return hostList, nil
}
