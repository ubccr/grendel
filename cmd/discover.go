package cmd

import (
	"fmt"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/ubccr/grendel/dhcp"
	"github.com/ubccr/grendel/tors"
	"github.com/urfave/cli"
)

func NewHostDiscoverCommand() cli.Command {
	return cli.Command{
		Name:        "discover",
		Usage:       "Auto-discover hosts from DHCP",
		Description: "Auto-discover hosts from DHCP",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "dhcp-port",
				Value: dhcpv4.ServerPort,
				Usage: "dhcp port to listen on",
			},
			cli.StringFlag{
				Name:  "listen-address",
				Value: "0.0.0.0",
				Usage: "address to listen on",
			},
			cli.StringFlag{
				Name:     "subnet",
				Required: true,
				Usage:    "subnet to use for auto ip assignment (/24)",
			},
			cli.StringFlag{
				Name:     "prefix",
				Required: true,
				Usage:    "hostname prefix",
			},
			cli.StringFlag{
				Name:  "suffix",
				Usage: "hostname suffix",
			},
			cli.StringFlag{
				Name:     "nodeset",
				Required: true,
				Usage:    "nodeset pattern",
			},
			cli.StringFlag{
				Name:  "switch-api-endpoint",
				Usage: "switch api endpoint",
			},
			cli.StringFlag{
				Name:  "switch-api-user",
				Usage: "switch api username",
			},
			cli.StringFlag{
				Name:  "switch-api-pass",
				Usage: "switch api password",
			},
		},
		Action: runHostDiscover,
	}
}

func runHostDiscover(c *cli.Context) error {
	listenAddress := c.String("listen-address")
	address := fmt.Sprintf("%s:%d", listenAddress, c.Int("dhcp-port"))

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

	return dhcp.RunDiscovery(DB, address, c.String("prefix"), c.String("suffix"), c.String("nodeset"), subnet, netmask, switchClient)
}
