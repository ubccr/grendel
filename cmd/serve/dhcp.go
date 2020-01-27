package serve

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/dhcp"
	"github.com/ubccr/grendel/logger"
)

func init() {
	dhcpCmd.PersistentFlags().String("dhcp-listen", "0.0.0.0:67", "address to listen on")
	viper.BindPFlag("dhcp.listen", dhcpCmd.PersistentFlags().Lookup("dhcp-listen"))
	dhcpCmd.PersistentFlags().String("dhcp-lease-time", "24h", "default lease time")
	viper.BindPFlag("dhcp.lease_time", dhcpCmd.PersistentFlags().Lookup("dhcp-lease-time"))
	dhcpCmd.PersistentFlags().StringSlice("dhcp-dns-servers", []string{}, "dns servers list")
	viper.BindPFlag("dhcp.dns_servers", dhcpCmd.PersistentFlags().Lookup("dhcp-dns-servers"))
	dhcpCmd.PersistentFlags().StringSlice("dhcp-domain-search", []string{}, "domain name search list")
	viper.BindPFlag("dhcp.domain_search", dhcpCmd.PersistentFlags().Lookup("dhcp-domain-search"))
	dhcpCmd.PersistentFlags().Int("dhcp-mtu", 1500, "default mtu")
	viper.BindPFlag("dhcp.mtu", dhcpCmd.PersistentFlags().Lookup("dhcp-mtu"))
	dhcpCmd.PersistentFlags().Bool("dhcp-proxy-only", false, "only run boot proxy")
	viper.BindPFlag("dhcp.proxy_only", dhcpCmd.PersistentFlags().Lookup("dhcp-proxy-only"))

	serveCmd.AddCommand(dhcpCmd)
}

var (
	dhcpLog = logger.GetLogger("DHCP")
	dhcpCmd = &cobra.Command{
		Use:   "dhcp",
		Short: "Run DHCP server",
		Long:  `Run DHCP server`,
		RunE: func(command *cobra.Command, args []string) error {
			ctx, _ := NewInterruptContext()
			return serveDHCP(ctx)
		},
	}
)

func serveDHCP(ctx context.Context) error {
	srv, err := dhcp.NewServer(DB, viper.GetString("dhcp.listen"))
	if err != nil {
		return err
	}

	address := viper.GetString("provision.listen")
	ipStr, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	provisionIP := net.ParseIP(ipStr)
	if provisionIP == nil || provisionIP.To4() == nil {
		return fmt.Errorf("Invalid Provision IPv4 address: %s", ipStr)
	}

	if provisionIP.To4().Equal(net.IPv4zero) {
		// Assume we're running on same server as provision?
		provisionIP = srv.ServerAddress
	}

	srv.HTTPPort, err = strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	srv.HTTPScheme = viper.GetString("provision.scheme")
	if srv.HTTPScheme == "" {
		srv.HTTPScheme = "http"
	}

	if viper.IsSet("provision.hostname") {
		srv.Hostname = viper.GetString("provision.hostname")
	}

	if srv.Hostname == "" && srv.HTTPScheme == "https" {
		hosts, err := net.LookupAddr(provisionIP.String())
		if err == nil && len(hosts) > 0 {
			fqdn := hosts[0]
			srv.Hostname = strings.TrimSuffix(fqdn, ".")
		}
	}

	if srv.Hostname == "" {
		srv.Hostname = provisionIP.String()
	}

	dhcpLog.Infof("Base URL for ipxe: %s://%s:%d", srv.HTTPScheme, srv.Hostname, srv.HTTPPort)

	if viper.IsSet("dhcp.dns_servers") {
		srv.DNSServers = make([]net.IP, 0)
		for _, arg := range viper.GetStringSlice("dhcp.dns_servers") {
			dnsip := net.ParseIP(arg)
			if dnsip.To4() == nil {
				return fmt.Errorf("Invalid dns server ip address: %s", arg)
			}
			srv.DNSServers = append(srv.DNSServers, dnsip)
		}
		dhcpLog.Infof("Using DNS servers: %v", srv.DNSServers)
	}

	if viper.IsSet("dhcp.domain_search") {
		srv.DomainSearchList = viper.GetStringSlice("dhcp.domain_search")
		dhcpLog.Infof("Using Domain Search List: %v", srv.DomainSearchList)
	}

	leaseTime, err := time.ParseDuration(viper.GetString("dhcp.lease_time"))
	if err != nil {
		return err
	}
	srv.LeaseTime = leaseTime
	srv.MTU = viper.GetInt("dhcp.mtu")
	dhcpLog.Infof("Default lease time: %s", srv.LeaseTime)
	dhcpLog.Infof("Default mtu: %d", srv.MTU)

	srv.ProxyOnly = viper.GetBool("dhcp.proxy_only")
	if srv.ProxyOnly {
		dhcpLog.Infof("Running in ProxyOnly mode")
	}

	if err := srv.Serve(ctx); err != nil {
		return err
	}

	return nil
}
