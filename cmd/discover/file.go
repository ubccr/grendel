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
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
)

var (
	fileCmd = &cobra.Command{
		Use:   "file",
		Short: "Discover hosts from file",
		Long:  `Discover hosts from file`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			for _, name := range args {
				file, err := os.Open(name)
				if err != nil {
				}
				defer file.Close()

				cmd.Log.Infof("Processing file: %s", name)
				err = discoverFromTSV(file)
				if err != nil {
					return err
				}

				cmd.Log.Infof("Successfully processed hosts from: %s", name)
			}

			return nil
		},
	}
)

func init() {
	discoverCmd.AddCommand(fileCmd)
}

func discoverFromTSV(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		cols := strings.Split(scanner.Text(), "\t")
		if len(cols) < 3 {
			return fmt.Errorf("Invalid record format. Must be at least 3 cols name|mac|ip: %s", line)
		}

		hwaddr, err := net.ParseMAC(cols[1])
		if err != nil {
			return fmt.Errorf("Malformed hardware address: %s", cols[0])
		}
		ipaddr := net.ParseIP(cols[2])
		if ipaddr.To4() == nil {
			return fmt.Errorf("Invalid IPv4 address: %v", cols[1])
		}

		fqdn := ""
		if len(cols) > 3 {
			fqdn = cols[3]
		}

		addNic(cols[0], fqdn, hwaddr, ipaddr, false)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
