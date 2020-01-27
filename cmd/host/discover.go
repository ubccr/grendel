package host

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/dhcp"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/tors"
)

var (
	mappingFile  string
	subnetStr    string
	bmcSubnetStr string
	discoverCmd  = &cobra.Command{
		Use:   "discover",
		Short: "Auto-discover hosts",
		Long:  `Auto-discover hosts`,
		RunE: func(command *cobra.Command, args []string) error {
			var switchClient tors.NetworkSwitch

			if viper.GetString("discovery.endpoint") != "" {
				sw, err := tors.NewDellOS10(viper.GetString("discovery.endpoint"), viper.GetString("discovery.user"), viper.GetString("discovery.password"), "", true)
				if err != nil {
					return err
				}

				switchClient = sw
			}

			netmask := net.IPv4Mask(255, 255, 255, 0)
			subnet := net.ParseIP(subnetStr)
			if subnet == nil || subnet.To4() == nil {
				return fmt.Errorf("Invalid IPv4 subnet address: %s", subnetStr)
			}
			bmcSubnet := net.ParseIP(bmcSubnetStr)
			if bmcSubnet != nil && subnet.To4() == nil {
				return fmt.Errorf("Invalid IPv4 bmc subnet address: %s", bmcSubnetStr)
			}

			if mappingFile != "" {
				if switchClient == nil {
					return fmt.Errorf("Switch is required when discovering hosts from file")
				}

				return discoverFromFile(mappingFile, viper.GetString("discovery.domain"), subnet, bmcSubnet, netmask, switchClient)
			}

			if len(args) == 0 {
				return errors.New("Please provide a nodeset")
			}

			return dhcp.RunDiscovery(viper.GetString("discovery.listen"), strings.Join(args, ","), subnet, netmask, switchClient)
		},
	}
)

func init() {
	discoverCmd.Flags().StringP("listen", "l", "0.0.0.0:67", "address to run discovery DHCP server")
	viper.BindPFlag("discovery.listen", discoverCmd.Flags().Lookup("listen"))
	discoverCmd.Flags().StringP("domain", "d", "", "domain name")
	viper.BindPFlag("discovery.domain", discoverCmd.Flags().Lookup("domain"))
	discoverCmd.Flags().StringP("user", "u", "", "switch api username")
	viper.BindPFlag("discovery.user", discoverCmd.Flags().Lookup("user"))
	discoverCmd.Flags().StringP("password", "p", "", "switch api password")
	viper.BindPFlag("discovery.password", discoverCmd.Flags().Lookup("password"))
	discoverCmd.Flags().StringP("endpoint", "e", "", "switch api endpoint")
	viper.BindPFlag("discovery.endpoint", discoverCmd.Flags().Lookup("endpoint"))

	discoverCmd.Flags().StringVarP(&mappingFile, "mapping", "m", "", "hostname to portnumber mapping file")
	discoverCmd.Flags().StringVarP(&subnetStr, "subnet", "s", "", "subnet to use for auto ip assignment (/24)")
	discoverCmd.Flags().StringVarP(&bmcSubnetStr, "bmc-subnet", "b", "", "subnet for bmc")

	discoverCmd.MarkFlagRequired("subnet")
	hostCmd.AddCommand(discoverCmd)
}

func discoverFromFile(file, domain string, subnet, bmcSubnet net.IP, netmask net.IPMask, switchClient tors.NetworkSwitch) error {
	log := logger.GetLogger("DISCOVER")

	macTable, err := switchClient.GetMACTable()
	if err != nil {
		return err
	}

	reader, err := os.Open(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	hosts := make([]*model.Host, 0)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "\t")

		port, err := strconv.Atoi(cols[1])
		if err != nil {
			return err
		}

		entries := macTable.Port(port)
		if len(entries) == 0 {
			log.Errorf("No port entries found on switch for node: %s port: %d", cols[0], port)
			continue
		}

		if len(entries) <= 1 {
			log.Warnf("Only found 1 port entry. Missing BMC?: %s port: %d", cols[0], port)
		}

		host := &model.Host{
			Name:       cols[0],
			Provision:  true,
			Firmware:   firmware.SNPONLY,
			Interfaces: make([]*model.NetInterface, 0),
		}

		for _, entry := range entries {
			vlanID, err := strconv.Atoi(strings.Replace(entry.VLAN, "vlan", "", -1))
			if err != nil {
				log.Errorf("Failed to parse vlan id for node %s: %s - %s", cols[0], entry.VLAN, entry.Ifname)
				continue
			}

			if vlanID > 1000 && vlanID < 2000 {
				ip := subnet.Mask(netmask)
				ip[2] += uint8(vlanID - 1000)
				ip[3] += uint8(port)
				nic := &model.NetInterface{
					MAC:  entry.MAC,
					IP:   ip,
					FQDN: fmt.Sprintf("%s.%s", cols[0], domain),
				}
				host.Interfaces = append(host.Interfaces, nic)
				continue
			} else if vlanID > 3000 && vlanID < 4000 && bmcSubnet != nil {
				ip := bmcSubnet.Mask(netmask)
				ip[2] += uint8(vlanID - 3000)
				ip[3] += uint8(port)
				nic := &model.NetInterface{
					MAC:  entry.MAC,
					IP:   ip,
					FQDN: fmt.Sprintf("%s.%s", strings.Replace(cols[0], "cpn", "bmc", -1), domain),
					BMC:  true,
				}
				host.Interfaces = append(host.Interfaces, nic)
				continue
			}

			log.Errorf("Unknown vlan id range: %s - %s %s", cols[0], entry.VLAN, entry.Ifname)
		}

		hosts = append(hosts, host)
	}

	b, err := json.Marshal(hosts)
	if err != nil {
		return err
	}

	os.Stdout.Write(b)

	return nil
}
