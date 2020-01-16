package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/ubccr/grendel/dhcp"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/tors"
	"github.com/urfave/cli/v2"
)

func NewHostDiscoverCommand() *cli.Command {
	return &cli.Command{
		Name:        "discover",
		Usage:       "Auto-discover hosts from DHCP",
		Description: "Auto-discover hosts from DHCP",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "dhcp-port",
				Value: dhcpv4.ServerPort,
				Usage: "dhcp port to listen on",
			},
			&cli.StringFlag{
				Name:  "listen-address",
				Value: "0.0.0.0",
				Usage: "address to listen on",
			},
			&cli.StringFlag{
				Name:  "from-file",
				Usage: "use hostname to portnumber mapping file instead of DHCP",
			},
			&cli.StringFlag{
				Name:     "subnet",
				Required: true,
				Usage:    "subnet to use for auto ip assignment (/24)",
			},
			&cli.StringFlag{
				Name:  "bmc-subnet",
				Usage: "subnet for bmc",
			},
			&cli.StringFlag{
				Name:  "domain",
				Usage: "domain",
			},
			&cli.StringFlag{
				Name:  "nodeset",
				Usage: "node set",
			},
			&cli.StringFlag{
				Name:  "switch-api-endpoint",
				Usage: "switch api endpoint",
			},
			&cli.StringFlag{
				Name:  "switch-api-user",
				Usage: "switch api username",
			},
			&cli.StringFlag{
				Name:  "switch-api-pass",
				Usage: "switch api password",
			},
		},
		Action: runHostDiscover,
	}
}

func runHostDiscover(c *cli.Context) error {
	var switchClient tors.NetworkSwitch

	if c.IsSet("switch-api-endpoint") {
		sw, err := tors.NewDellOS10(c.String("switch-api-endpoint"), c.String("switch-api-user"), c.String("switch-api-pass"), "", true)
		if err != nil {
			return err
		}

		switchClient = sw
	}

	netmask := net.IPv4Mask(255, 255, 255, 0)
	subnet := net.ParseIP(c.String("subnet"))
	if subnet == nil || subnet.To4() == nil {
		return fmt.Errorf("Invalid IPv4 subnet address: %s", c.String("subnet"))
	}
	bmcSubnet := net.ParseIP(c.String("bmc-subnet"))
	if bmcSubnet != nil && subnet.To4() == nil {
		return fmt.Errorf("Invalid IPv4 bmc subnet address: %s", c.String("bmc-subnet"))
	}

	if c.IsSet("from-file") {
		if switchClient == nil {
			return fmt.Errorf("Switch is required when discovering hosts from file")
		}

		return discoverFromFile(c.String("from-file"), c.String("domain"), subnet, bmcSubnet, netmask, switchClient)
	}

	listenAddress := c.String("listen-address")
	address := fmt.Sprintf("%s:%d", listenAddress, c.Int("dhcp-port"))

	return dhcp.RunDiscovery(address, c.String("nodeset"), subnet, netmask, switchClient)
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
