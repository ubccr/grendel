package cmd

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	log "github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/dhcp"
	"github.com/urfave/cli"
)

func NewDHCPCommand() cli.Command {
	return cli.Command{
		Name:        "dhcp",
		Usage:       "Start DHCP server",
		Description: "Start DHCP server",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "port",
				Value: dhcpv4.ServerPort,
				Usage: "dhcp port to listen on",
			},
			cli.StringFlag{
				Name:  "listen-address",
				Usage: "dhcp address to listen on",
			},
			cli.StringFlag{
				Name:  "static-leases",
				Usage: "static dhcp leases file",
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
		},
		Action: runDHCP,
	}
}

func runDHCP(c *cli.Context) error {
	listenAddress := c.GlobalString("listen-address")
	if c.IsSet("listen-address") {
		listenAddress = c.String("listen-address")
	}

	address := fmt.Sprintf("%s:%d", listenAddress, c.Int("port"))

	srv, err := dhcp.NewServer(address)
	if err != nil {
		return err
	}

	if c.GlobalString("cert") != "" && c.GlobalString("key") != "" {
		srv.HTTPScheme = "https"
		srv.HTTPPort = 443
		hosts, err := net.LookupAddr(srv.ServerAddress.String())
		if err == nil && len(hosts) > 0 {
			fqdn := hosts[0]
			srv.Hostname = strings.TrimSuffix(fqdn, ".")
			log.Infof("Using HTTPS for ipxe: %s://%s:%d", srv.HTTPScheme, srv.Hostname, srv.HTTPPort)
		} else {
			log.Warning("Failed to lookup server hostname for HTTPS. Using IP")
			log.Infof("Using HTTPS for ipxe: %s://%s:%d", srv.HTTPScheme, srv.ServerAddress, srv.HTTPPort)
		}
	}

	if c.IsSet("static-leases") {
		err := srv.LoadStaticLeases(c.String("static-leases"))
		if err != nil {
			return err
		}
	}

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

	return srv.Serve()
}
