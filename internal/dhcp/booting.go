// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package dhcp

import (
	"fmt"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/internal/firmware"
	"github.com/ubccr/grendel/pkg/model"
)

func (s *Server) bootingHandler4(host *model.Host, serverIP net.IP, req, resp *dhcpv4.DHCPv4) error {
	if !host.Provision {
		log.Infof("Host not set to provision: %s", req.ClientHWAddr.String())
		return nil
	}

	if !req.Options.Has(dhcpv4.OptionClientSystemArchitectureType) {
		log.Debugf("Ignoring packet - missing client system architecture type")
		return nil
	}

	userClass := ""
	if req.Options.Has(dhcpv4.OptionUserClassInformation) {
		userClass = string(req.Options.Get(dhcpv4.OptionUserClassInformation))
	}

	fwtype, err := firmware.DetectBuild(req.ClientArch(), userClass)
	log.Debugf("iPXE Firmware type detected: %s", fwtype.String())
	if err != nil {
		return fmt.Errorf("Failed to get PXE firmware from DHCP: %s", err)
	}

	log.WithFields(logrus.Fields{
		"mac":      req.ClientHWAddr.String(),
		"name":     host.Name,
		"firmware": fwtype.String(),
	}).Info("Got valid PXE boot request")
	log.Debugf(req.Summary())

	// This logic was adopted from pixiecore
	// https://github.com/danderson/netboot/tree/master/pixiecore
	// Written by @danderson
	switch fwtype {
	case firmware.UNDI:
		if !s.ProxyOnly {
			// If we're running both dhcp server and PXE Server then we need to
			// bail here to direct the PXE client over to port 4011 for the
			// bootfile. This is because we're running both dhcp and PXE on
			// same server
			return nil
		}

		log.Printf("UNDI telling PXE client to bypass all boot discovery")
		pxe := dhcpv4.OptionsFromList(dhcpv4.OptGeneric(dhcpv4.GenericOptionCode(6), []byte{8}))
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionVendorSpecificInformation, pxe.ToBytes()))

		resp.UpdateOption(dhcpv4.OptTFTPServerName(serverIP.String()))

		token, err := model.NewFirmwareToken(req.ClientHWAddr.String(), fwtype)
		if err != nil {
			return fmt.Errorf("UNDI failed to generated signed Firmware token")
		}
		resp.UpdateOption(dhcpv4.OptBootFileName(token))

	case firmware.IPXE:
		log.Printf("Found iPXE firmware telling PXE client to boot tftp")
		pxe := dhcpv4.OptionsFromList(dhcpv4.OptGeneric(dhcpv4.GenericOptionCode(6), []byte{8}))
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionVendorSpecificInformation, pxe.ToBytes()))

		resp.UpdateOption(dhcpv4.OptTFTPServerName(serverIP.String()))

		token, err := model.NewFirmwareToken(req.ClientHWAddr.String(), fwtype)
		if err != nil {
			return fmt.Errorf("iPXE firmware - failed to generated signed Firmware token")
		}
		endpoints := model.NewEndpoints(serverIP.String(), token)
		resp.UpdateOption(dhcpv4.OptBootFileName(endpoints.BootFileURL()))

	case firmware.EFI386, firmware.EFI64, firmware.SNPONLYarm64:
		log.Printf("EFI boot PXE client")
		if host.Firmware != 0 {
			log.Infof("Overriding firmware for host: %s", req.ClientHWAddr.String())
			fwtype = host.Firmware
		}
		resp.UpdateOption(dhcpv4.OptTFTPServerName(serverIP.String()))

		token, err := model.NewFirmwareToken(req.ClientHWAddr.String(), fwtype)
		if err != nil {
			return fmt.Errorf("EFI failed to generated signed Firmware token")
		}
		resp.UpdateOption(dhcpv4.OptBootFileName(token))

	case firmware.GRENDEL:
		// Chainload to HTTP
		token, err := model.NewBootToken(host.ID.String(), req.ClientHWAddr.String())
		if err != nil {
			return fmt.Errorf("Failed to generate signed boot token: %s", err)
		}

		endpoints := model.NewEndpoints(serverIP.String(), token)
		ipxeUrl := endpoints.IpxeURL()
		log.Debugf("BootFile iPXE script: %s", ipxeUrl)
		resp.UpdateOption(dhcpv4.OptBootFileName(ipxeUrl))

	default:
		return fmt.Errorf("unknown firmware type %d", fwtype)
	}

	return nil
}
