// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package dhcp

import (
	"fmt"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/rfc1035label"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/model"
)

func (s *Server) staticHandler4(host *model.Host, req, resp *dhcpv4.DHCPv4) error {
	nic := host.Interface(req.ClientHWAddr)
	if nic == nil {
		log.Warnf("invalid mac address for host: %s", req.ClientHWAddr)
		return nil
	}

	resp.YourIPAddr = nic.IP
	log.WithFields(logrus.Fields{
		"ip":           nic.IP.String(),
		"mac":          req.ClientHWAddr.String(),
		"name":         host.Name,
		"dhcp_message": req.MessageType().String(),
	}).Info("Found host")
	log.Debugf(req.Summary())

	if s.Netmask != nil {
		resp.UpdateOption(dhcpv4.OptSubnetMask(s.Netmask))
	}

	var router net.IP
	if s.RouterOctet4 > 0 {
		router = nic.IP.Mask(s.Netmask)
		router[3] += byte(s.RouterOctet4)
	} else if s.RouterIP != nil {
		router = make(net.IP, len(s.RouterIP))
		copy(router, s.RouterIP)
	}

	if router != nil {
		routers := []net.IP{router}
		resp.UpdateOption(dhcpv4.OptRouter(routers...))
	}

	resp.UpdateOption(dhcpv4.OptIPAddressLeaseTime(s.LeaseTime))

	if len(s.DNSServers) > 0 && req.IsOptionRequested(dhcpv4.OptionDomainNameServer) {
		resp.UpdateOption(dhcpv4.OptDNS(s.DNSServers...))
	}

	if req.IsOptionRequested(dhcpv4.OptionInterfaceMTU) {
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionInterfaceMTU, dhcpv4.Uint16(s.MTU).ToBytes()))
	}

	if len(nic.FQDN) > 0 {
		resp.UpdateOption(dhcpv4.OptHostName(nic.FQDN))
	}

	if len(s.DomainSearchList) > 0 {
		resp.UpdateOption(dhcpv4.OptDomainSearch(&rfc1035label.Labels{
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
		// Need to return NACK here. The client is asking for a different IP
		// address than what's configured in Grendel.
		msg := fmt.Sprintf("Requested IP address %v does not match address configured in Grendel: %v", requestedIP, nic.IP)
		log.Info(msg)
		resp.UpdateOption(dhcpv4.OptMessage(msg))
		resp.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeNak))
		return nil
	}

	if req.ClientIPAddr != nil && !req.ClientIPAddr.Equal(net.IPv4zero) {
		resp.ClientIPAddr = req.ClientIPAddr
	}

	resp.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
	return s.staticHandler4(host, req, resp)
}
