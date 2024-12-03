// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package discover

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/go-dhcpd-leases"
	"github.com/ubccr/grendel/cmd"
)

var (
	fileType string
	fileCmd  = &cobra.Command{
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
				if ".leases" == filepath.Ext(name) || "leases" == strings.ToLower(fileType) {
					err = discoverFromLeases(file)
				} else {
					// Default to parsing TSV?
					err = discoverFromTSV(file)
				}

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
	fileCmd.Flags().StringVar(&fileType, "type", "", "file type to parse")
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

func discoverFromLeases(reader io.Reader) error {
	hosts := leases.Parse(reader)
	if hosts == nil {
		return errors.New("No hosts found. Is this a dhcpd.leases file?")
	}

	for _, h := range hosts {
		names := strings.Split(h.ClientHostname, ".")
		addNic(names[0], h.ClientHostname, h.Hardware.MACAddr, h.IP, false)
	}

	return nil
}
