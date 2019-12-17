package cmd

import (
	"fmt"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/ubccr/grendel/dhcp"
	"github.com/ubccr/grendel/tor"
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

	var switchClient tor.NetworkSwitch

	if c.IsSet("switch-api-endpoint") {
		sw, err := tor.NewDellOS10(c.String("switch-api-endpoint"), c.String("switch-api-user"), c.String("switch-api-pass"), "", true)
		if err != nil {
			return err
		}

		switchClient = sw
	}

	return dhcp.RunDiscovery(DB, address, c.String("prefix"), c.String("nodeset"), switchClient)
}
