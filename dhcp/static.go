package dhcp

import (
	"fmt"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/rfc1035label"
	"github.com/ubccr/grendel/model"
)

func (s *Server) staticHandler4(host *model.Host, req, resp *dhcpv4.DHCPv4) error {
	nic := host.Interface(req.ClientHWAddr)
	if nic == nil {
		log.Warnf("StaticHandler4 invalid mac address for host: %s", req.ClientHWAddr)
		return nil
	}

	resp.YourIPAddr = nic.IP
	log.Infof("StaticHandler4 found IP address %s for MAC %s", nic.IP, req.ClientHWAddr.String())
	log.Debugf(req.Summary())

	// TODO make this configurable
	netmask := net.IPv4Mask(255, 255, 255, 0)
	resp.Options.Update(dhcpv4.OptSubnetMask(netmask))

	// Use 10.x.x.254 as the router
	router := nic.IP.Mask(netmask)
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

	if len(nic.FQDN) > 0 {
		resp.Options.Update(dhcpv4.OptHostName(nic.FQDN))
	}

	if len(s.DomainSearchList) > 0 {
		resp.Options.Update(dhcpv4.OptDomainSearch(&rfc1035label.Labels{
			Labels: s.DomainSearchList,
		}))
	}

	return nil
}

func (s *Server) staticAckHandler4(host *model.Host, req, resp *dhcpv4.DHCPv4) error {
	if req.ServerIPAddr != nil &&
		!req.ServerIPAddr.Equal(net.IPv4zero) &&
		!req.ServerIPAddr.Equal(s.ServerAddress) {
		return fmt.Errorf("requested ServerID does not match. Got %v, want %v", req.ServerIPAddr, s.ServerAddress)
	}

	if req.ServerIdentifier() != nil &&
		!req.ServerIdentifier().Equal(net.IPv4zero) &&
		!req.ServerIdentifier().Equal(s.ServerAddress) {
		return fmt.Errorf("requested Server Identifier does not match. Got %v, want %v", req.ServerIdentifier(), s.ServerAddress)
	}

	requestedIP := req.RequestedIPAddress()
	if requestedIP == nil || requestedIP.Equal(net.IPv4zero) {
		requestedIP = req.ClientIPAddr
	}

	nic := host.Interface(req.ClientHWAddr)
	if !nic.IP.Equal(requestedIP) {
		// TODO Need to return NACK here
		return fmt.Errorf("Requested IP address does not match this hardware address. Got %v, want %v", requestedIP, nic.IP)
	}

	if req.ClientIPAddr != nil && !req.ClientIPAddr.Equal(net.IPv4zero) {
		resp.ClientIPAddr = req.ClientIPAddr
	}

	resp.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
	return s.staticHandler4(host, req, resp)
}
