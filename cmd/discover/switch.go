// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package discover

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/tors"
)

var (
	mappingFile  string
	bmcSubnetStr string
	switchCmd    = &cobra.Command{
		Use:   "switch",
		Short: "Auto-discover hosts from switch",
		Long:  `Auto-discover hosts from switch`,
		RunE: func(command *cobra.Command, args []string) error {
			if subnetStr == "" && bmcSubnetStr == "" {
				return fmt.Errorf("Please provide a least one subnet (--subnet and/or --bmc-subnet)")
			}

			endpoint := viper.GetString("discovery.endpoint")

			var switchClient tors.NetworkSwitch
			var err error
			if strings.HasPrefix(endpoint, "http") {
				switchClient, err = tors.NewDellOS10(endpoint, viper.GetString("discovery.user"), viper.GetString("discovery.password"), "", true)
			} else {
				switchClient, err = tors.NewGeneric(endpoint, "public")
			}

			if err != nil {
				return err
			}

			bmcSubnet := net.IPv4(0, 0, 0, 0)
			if bmcSubnetStr != "" {
				bmcSubnet = net.ParseIP(bmcSubnetStr)
				if bmcSubnet != nil && bmcSubnet.To4() == nil {
					return fmt.Errorf("Invalid IPv4 BMC subnet address: %s", bmcSubnetStr)
				}
			}

			// TODO make this configurable?
			netmask := net.IPv4Mask(255, 255, 0, 0)

			return discoverFromSwitch(mappingFile, viper.GetString("discovery.domain"), subnet, bmcSubnet, netmask, switchClient)
		},
	}
)

func init() {
	switchCmd.Flags().StringP("user", "u", "", "switch api username")
	viper.BindPFlag("discovery.user", switchCmd.Flags().Lookup("user"))
	switchCmd.Flags().StringP("password", "p", "", "switch api password")
	viper.BindPFlag("discovery.password", switchCmd.Flags().Lookup("password"))
	switchCmd.Flags().StringP("endpoint", "e", "", "switch api endpoint")
	viper.BindPFlag("discovery.endpoint", switchCmd.Flags().Lookup("endpoint"))

	switchCmd.Flags().StringVarP(&mappingFile, "mapping", "m", "", "hostname to portnumber mapping file")
	switchCmd.Flags().StringVarP(&bmcSubnetStr, "bmc-subnet", "b", "", "subnet for bmc")

	switchCmd.MarkFlagRequired("endpoint")
	switchCmd.MarkFlagRequired("mapping")

	discoverCmd.AddCommand(switchCmd)
}

func discoverFromSwitch(file, domain string, subnet, bmcSubnet net.IP, netmask net.IPMask, switchClient tors.NetworkSwitch) error {

	reader, err := os.Open(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	macTable, err := switchClient.GetMACTable()
	if err != nil {
		return err
	}

	log.Debugf("MAC Table: %s", macTable)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "\t")
		if len(cols) != 2 {
			log.Warnf("Invalid mapping format")
			break
		}

		hostName := cols[0]

		port, err := strconv.Atoi(cols[1])
		if err != nil {
			return err
		}

		entries := macTable.Port(port)
		if len(entries) == 0 {
			log.Warnf("No port entries found on switch for node: %s port: %d", hostName, port)
			continue
		}

		if len(entries) <= 1 && !bmcSubnet.To4().Equal(net.IPv4zero) && !subnet.To4().Equal(net.IPv4zero) {
			log.Warnf("Only found 1 port entry. Missing BMC?: %s port: %d", hostName, port)
		}

		for _, entry := range entries {
			vlanID, err := strconv.Atoi(strings.Replace(entry.VLAN, "vlan", "", -1))
			if err != nil {
				log.Errorf("Failed to parse vlan id for node %s: %s - %s", hostName, entry.VLAN, entry.Ifname)
				continue
			}

			if vlanID < 1000 {
				log.Errorf("Unknown vlan id: %s - %s %s", hostName, entry.VLAN, entry.Ifname)
			}

			vlanClass := 1000 * int(vlanID/1000)
			switchID := vlanID - vlanClass
			log.Debugf("Found vlanID: %d vlanClass: %d switchID: %d", vlanID, vlanClass, switchID)

			ip := subnet.Mask(netmask)
			isBMC := false
			fqdn := fmt.Sprintf("%s.%s", hostName, domain)

			if (subnet.To4().Equal(net.IPv4zero) && !bmcSubnet.To4().Equal(net.IPv4zero)) ||
				(!subnet.To4().Equal(net.IPv4zero) && !bmcSubnet.To4().Equal(net.IPv4zero) && vlanClass == 3000) {
				ip = bmcSubnet.Mask(netmask)
				isBMC = true
				fqdn = strings.Replace(fqdn, "cpn", "bmc", -1)
			}

			ip[2] += uint8(switchID)
			ip[3] += uint8(port)
			addNic(hostName, fqdn, entry.MAC, ip, isBMC)
			continue

		}
	}

	return nil
}
