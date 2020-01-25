package cmd

import (
	"fmt"

	"github.com/ubccr/grendel/dns"
	"github.com/urfave/cli/v2"
)

func NewDNSCommand() *cli.Command {
	return &cli.Command{
		Name:        "dns",
		Usage:       "Start DNS server",
		Description: "Start DNS server",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "dns-port",
				Value: 53,
				Usage: "dns port to listen on",
			},
			&cli.IntFlag{
				Name:  "dns-ttl",
				Value: 86400,
				Usage: "dns ttl for records",
			},
			&cli.StringFlag{
				Name:  "listen-address",
				Value: "0.0.0.0",
				Usage: "address to listen on",
			},
		},
		Action: runDNS,
	}
}

func runDNS(c *cli.Context) error {
	listenAddress := c.String("listen-address")

	address := fmt.Sprintf("%s:%d", listenAddress, c.Int("dns-port"))

	dnsServer, err := dns.NewServer(DB, address, c.Int("dns-ttl"))
	if err != nil {
		return err
	}

	return dnsServer.Serve()
}
