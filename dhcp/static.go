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

func (s *Server) setZTD(host *model.Host, nic *model.NetInterface, serverIP net.IP, resp *dhcpv4.DHCPv4) {
	if !host.Provision {
		// Skip if host not set to provision
		return
	}

	if host.HasTags("dellbmp") {
		// Dell Bare Metal Provisioning (BMP) for FTOS
		// See: https://i.dell.com/sites/doccontent/shared-content/Documents/Bare_Metal_Provisioning.pdf

		log.WithFields(logrus.Fields{
			"ip":   nic.AddrString(),
			"name": host.Name,
		}).Info("Host tagged with Dell BMP. Setting FTOS image URL and config dhcp options")

		token, _ := model.NewBootToken(host.ID.String(), nic.MAC.String())
		endpoints := model.NewEndpoints(serverIP.String(), token)

		imageURL := endpoints.KernelURL()
		log.Debugf("Dell FTOS - image URL: %s", imageURL)
		resp.UpdateOption(dhcpv4.OptBootFileName(imageURL))

		configURL := endpoints.KickstartURL()
		log.Debugf("Dell FTOS - PXELinuxConfigFile: %s", configURL)
		resp.UpdateOption(dhcpv4.Option{Code: dhcpv4.OptionPXELinuxConfigFile, Value: dhcpv4.String(configURL)})
	}

	if host.HasTags("dellztd") {
		// Dell Zero-touch deployment (ZTD) for DellOS10
		// See: https://www.dell.com/support/manuals/en-in/networking-mx7116n/smartfabric-os-user-guide-10-5-0/dell-emc-smartfabric-os10-zero-touch-deployment?guid=guid-95ca07a2-2bcb-4ea2-84ef-ef9d11a4fa0e&lang=en-us

		log.WithFields(logrus.Fields{
			"ip":   nic.AddrString(),
			"name": host.Name,
		}).Info("Host tagged with Dell ZTD. Setting ZTD provision URL dhcp option")

		token, _ := model.NewBootToken(host.ID.String(), nic.MAC.String())
		endpoints := model.NewEndpoints(serverIP.String(), token)

		provisionURL := endpoints.KickstartURL()
		log.Debugf("Dell ZTD provision-url: %s", provisionURL)
		resp.UpdateOption(dhcpv4.Option{Code: dhcpv4.GenericOptionCode(240), Value: dhcpv4.String(provisionURL)})
	}
}

func (s *Server) staticHandler4(host *model.Host, serverIP net.IP, req, resp *dhcpv4.DHCPv4) error {
	nic := host.Interface(req.ClientHWAddr)
	if nic == nil {
		log.Warnf("invalid mac address for host: %s", req.ClientHWAddr)
		return nil
	}

	log.WithFields(logrus.Fields{
		"ip":           nic.AddrString(),
		"mac":          req.ClientHWAddr.String(),
		"name":         host.Name,
		"dhcp_message": req.MessageType().String(),
	}).Info("Found host")
	log.Debugf(req.Summary())

	resp.YourIPAddr = nic.ToStdAddr()
	resp.UpdateOption(dhcpv4.OptSubnetMask(nic.Netmask()))
	resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionInterfaceMTU, dhcpv4.Uint16(nic.InterfaceMTU()).ToBytes()))
	resp.UpdateOption(dhcpv4.OptIPAddressLeaseTime(s.LeaseTime))

	routerIP := nic.Gateway()
	if routerIP.IsValid() {
		routers := []net.IP{net.IP(routerIP.AsSlice())}
		resp.UpdateOption(dhcpv4.OptRouter(routers...))
	}

	dnsServers := nic.DNS()
	if len(dnsServers) > 0 {
		resp.UpdateOption(dhcpv4.OptDNS(dnsServers...))
	}

	if nic.FQDN != "" {
		resp.UpdateOption(dhcpv4.OptHostName(nic.FQDN))
	}

	domainSearch := nic.DomainSearch()
	if len(domainSearch) > 0 {
		resp.UpdateOption(dhcpv4.OptDomainSearch(&rfc1035label.Labels{
			Labels: domainSearch,
		}))
	}

	s.setZTD(host, nic, serverIP, resp)

	return nil
}

func (s *Server) staticAckHandler4(host *model.Host, serverIP net.IP, req, resp *dhcpv4.DHCPv4) error {
	if req.ServerIPAddr != nil &&
		!req.ServerIPAddr.Equal(net.IPv4zero) &&
		!req.ServerIPAddr.Equal(serverIP) {
		return fmt.Errorf("requested ServerID does not match. Got %v, want %v", req.ServerIPAddr, serverIP)
	}

	if req.ServerIdentifier() != nil &&
		!req.ServerIdentifier().Equal(net.IPv4zero) &&
		!req.ServerIdentifier().Equal(serverIP) {
		return fmt.Errorf("requested Server Identifier does not match. Got %v, want %v", req.ServerIdentifier(), serverIP)
	}

	requestedIP := req.RequestedIPAddress()
	if requestedIP == nil || requestedIP.Equal(net.IPv4zero) {
		requestedIP = req.ClientIPAddr
	}

	nic := host.Interface(req.ClientHWAddr)
	if !nic.ToStdAddr().Equal(requestedIP) {
		// Need to return NACK here. The client is asking for a different IP
		// address than what's configured in Grendel.
		msg := fmt.Sprintf("Requested IP address %v does not match address configured in Grendel: %v", requestedIP, nic.AddrString())
		log.Info(msg)
		resp.UpdateOption(dhcpv4.OptMessage(msg))
		resp.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeNak))
		return nil
	}

	if req.ClientIPAddr != nil && !req.ClientIPAddr.Equal(net.IPv4zero) {
		resp.ClientIPAddr = req.ClientIPAddr
	}

	resp.UpdateOption(dhcpv4.OptMessageType(dhcpv4.MessageTypeAck))
	return s.staticHandler4(host, serverIP, req, resp)
}
