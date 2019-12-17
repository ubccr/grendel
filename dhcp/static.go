package dhcp

import (
	"fmt"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
)

func (s *Server) staticHandler4(req, resp *dhcpv4.DHCPv4) error {
	host, err := s.DB.GetHost(req.ClientHWAddr.String())
	if err != nil {
		return err
	}

	resp.YourIPAddr = host.IP
	log.Infof("Found IP address %s for MAC %s", host.IP, req.ClientHWAddr.String())

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

	host, err := s.DB.GetHost(req.ClientHWAddr.String())
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
