package cmd

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/ubccr/grendel/dhcp"
	"github.com/ubccr/grendel/logger"
	"github.com/urfave/cli"
)

func NewDHCPCommand() cli.Command {
	return cli.Command{
		Name:        "dhcp",
		Usage:       "Start DHCP server",
		Description: "Start DHCP server",
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
				Name:  "http-scheme",
				Value: "http",
				Usage: "http scheme",
			},
			cli.IntFlag{
				Name:  "http-port",
				Value: 80,
				Usage: "http port",
			},
			cli.IntFlag{
				Name:  "pxe-port",
				Value: 4011,
				Usage: "pxe port",
			},
			cli.StringFlag{
				Name:  "hostname",
				Usage: "Hostname",
			},
			cli.StringFlag{
				Name:  "lease-time",
				Value: "24h",
				Usage: "dhcp lease time duration",
			},
			cli.StringSliceFlag{
				Name:  "dns-server",
				Usage: "dns server IP addresses",
			},
			cli.IntFlag{
				Name:  "mtu",
				Value: 1500,
				Usage: "dhcp interface MTU",
			},
			cli.BoolFlag{
				Name:  "disable-pxe",
				Usage: "Disable PXE server",
			},
		},
		Action: runDHCP,
	}
}

func runDHCP(c *cli.Context) error {
	log := logger.GetLogger("DHCP")

	listenAddress := c.String("listen-address")
	address := fmt.Sprintf("%s:%d", listenAddress, c.Int("dhcp-port"))

	srv, err := dhcp.NewServer(DB, address)
	if err != nil {
		return err
	}

	srv.HTTPScheme = c.String("http-scheme")
	srv.HTTPPort = c.Int("http-port")

	if srv.HTTPScheme == "https" && srv.HTTPPort == 80 {
		srv.HTTPPort = 443
	}

	srv.Hostname = c.String("hostname")

	if srv.Hostname == "" && srv.HTTPScheme == "https" {
		hosts, err := net.LookupAddr(srv.ServerAddress.String())
		if err == nil && len(hosts) > 0 {
			fqdn := hosts[0]
			srv.Hostname = strings.TrimSuffix(fqdn, ".")
		}
	}

	if srv.Hostname == "" {
		srv.Hostname = srv.ServerAddress.String()
	}

	log.Infof("Base URL for ipxe: %s://%s:%d", srv.HTTPScheme, srv.Hostname, srv.HTTPPort)

	if c.IsSet("dns-server") {
		srv.DNSServers = make([]net.IP, 0)
		for _, arg := range c.StringSlice("dns-server") {
			dnsip := net.ParseIP(arg)
			if dnsip.To4() == nil {
				return fmt.Errorf("Invalid dns server ip address: %s", arg)
			}
			srv.DNSServers = append(srv.DNSServers, dnsip)
		}
		log.Infof("Using DNS servers: %v", srv.DNSServers)
	}

	leaseTime, err := time.ParseDuration(c.String("lease-time"))
	if err != nil {
		return err
	}
	srv.LeaseTime = leaseTime
	srv.MTU = c.Int("mtu")

	srv.PXEPort = c.Int("pxe-port")
	srv.ServePXE = !c.Bool("disable-pxe")

	return srv.Serve()
}
