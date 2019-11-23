package dhcp

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/insomniacslk/dhcp/dhcpv4"
	log "github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/model"
)

func (s *Server) staticHandler4(req, resp *dhcpv4.DHCPv4) error {
	host, err := s.GetHost(req.ClientHWAddr.String())
	if err != nil {
		return err
	}

	resp.YourIPAddr = host.IP
	log.Debugf("found IP address %s for MAC %s", host.IP, req.ClientHWAddr.String())

	// TODO make this configurable
	netmask := net.IPv4Mask(255, 255, 255, 0)
	resp.Options.Update(dhcpv4.OptSubnetMask(netmask))

	// Use 10.x.x.254 as the router
	router := host.IP.Mask(netmask)
	router[3] += 254

	routers := []net.IP{router}
	resp.Options.Update(dhcpv4.OptRouter(routers...))

	resp.Options.Update(dhcpv4.OptIPAddressLeaseTime(s.LeaseTime))

	if len(s.DNSServers) > 0 && req.IsOptionRequested(dhcpv4.OptionDomainNameServer) {
		resp.Options.Update(dhcpv4.OptDNS(s.DNSServers...))
	}

	if req.IsOptionRequested(dhcpv4.OptionInterfaceMTU) {
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionInterfaceMTU, dhcpv4.Uint16(s.MTU).ToBytes()))
	}

	if len(host.FQDN) > 0 {
		resp.Options.Update(dhcpv4.OptHostName(host.FQDN))
	}

	return nil
}

func (s *Server) staticAckHandler4(req, resp *dhcpv4.DHCPv4) error {
	if req.ServerIPAddr != nil &&
		!req.ServerIPAddr.Equal(net.IPv4zero) &&
		!req.ServerIPAddr.Equal(s.ServerAddress) {
		// This request is not for us, drop it.
		return fmt.Errorf("requested server ID does not match this server's ID. Got %v, want %v", req.ServerIPAddr, s.ServerAddress)
	}

	host, err := s.GetHost(req.ClientHWAddr.String())
	if err != nil {
		return err
	}

	requestedIP := req.RequestedIPAddress()
	if requestedIP == nil || requestedIP.Equal(net.IPv4zero) {
		requestedIP = req.ClientIPAddr
	}

	if !host.IP.Equal(requestedIP) {
		return fmt.Errorf("Requested IP address does not match this hardware address. Got %v, want %v", requestedIP, host.IP)
	}

	resp.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
	return s.staticHandler4(req, resp)
}

func (s *Server) GetHost(mac string) (*model.Host, error) {
	host, ok := s.StaticLeases[mac]

	if !ok {
		return host, fmt.Errorf("MAC address %s is unknown", mac)
	}

	return host, nil
}

func (s *Server) LoadStaticLeases(filename string) error {
	log.Infof("Reading static leases from file: %s", filename)
	s.StaticLeases = make(map[string]*model.Host)

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "\t")
		hwaddr, err := net.ParseMAC(cols[0])
		if err != nil {
			return fmt.Errorf("Malformed hardware address: %s", cols[0])
		}
		ipaddr := net.ParseIP(cols[1])
		if ipaddr.To4() == nil {
			return fmt.Errorf("Invalid IPv4 address: %v", cols[1])
		}

		host := &model.Host{MAC: hwaddr, IP: ipaddr}

		if len(cols) > 2 {
			host.FQDN = cols[2]
		}

		s.StaticLeases[hwaddr.String()] = host
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(s.StaticLeases) > 0 {
		s.ProxyOnly = false
		log.Infof("Using %d static leases from %s", len(s.StaticLeases), filename)
	}

	return nil
}
