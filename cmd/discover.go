package cmd

import (
	"fmt"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/ubccr/grendel/dhcp"
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
				Name:     "prefix",
				Required: true,
				Usage:    "hostname prefix",
			},
			cli.StringFlag{
				Name:     "nodeset",
				Required: true,
				Usage:    "nodeset pattern",
			},
		},
		Action: runHostDiscover,
	}
}

func runHostDiscover(c *cli.Context) error {
	listenAddress := c.String("listen-address")
	address := fmt.Sprintf("%s:%d", listenAddress, c.Int("dhcp-port"))

	return dhcp.RunDiscovery(DB, address, c.String("prefix"), c.String("nodeset"))
}
