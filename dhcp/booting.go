package dhcp

import (
	"fmt"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/model"
)

func (s *Server) bootingHandler4(req, resp *dhcpv4.DHCPv4) error {
	if !req.Options.Has(dhcpv4.OptionClientSystemArchitectureType) {
		log.Debugf("BootHandler4 ignoring packet - missing client system architecture type")
		return nil
	}

	userClass := ""
	if req.Options.Has(dhcpv4.OptionUserClassInformation) {
		userClass = string(req.Options.Get(dhcpv4.OptionUserClassInformation))
	}

	fwtype, err := firmware.DetectBuild(req.ClientArch(), userClass)
	if err != nil {
		return fmt.Errorf("Failed to get PXE firmware from DHCP: %s", err)
	}

	log.Infof("Got valid request to boot %s %d", req.ClientIPAddr, fwtype)

	switch fwtype {
	case firmware.UNDI:
		if !s.ProxyOnly {
			// If we're running both dhcp server and PXE Server then we need to
			// bail here to direct the PXE client over to port 4011 for the
			// bootfile. This is because we're running both dhcp and PXE on
			// same server
			return nil
		}

		// This is completely standard PXE: we tell the PXE client to
		// bypass all the boot discovery rubbish that PXE supports,
		// and just load a file from TFTP.
		log.Printf("UNDI telling PXE client to bypass all boot discovery")

		pxe := dhcpv4.OptionsFromList(dhcpv4.OptGeneric(dhcpv4.GenericOptionCode(6), []byte{8}))
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionVendorSpecificInformation, pxe.ToBytes()))

		resp.UpdateOption(dhcpv4.OptTFTPServerName(s.ServerAddress.String()))

		token, err := model.NewFirmwareToken(req.ClientHWAddr.String(), fwtype)
		if err != nil {
			return fmt.Errorf("UNDI failed to generated signed Firmware token")
		}
		resp.UpdateOption(dhcpv4.OptBootFileName(token))

	case firmware.IPXE:
		log.Printf("Found iPXE firmware telling PXE client to boot tftp")
		// Almost standard PXE, but the boot filename needs to be a URL.

		// PXE Boot Server Discovery Control - bypass, just boot from filename.
		pxe := dhcpv4.OptionsFromList(dhcpv4.OptGeneric(dhcpv4.GenericOptionCode(6), []byte{8}))
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionVendorSpecificInformation, pxe.ToBytes()))

		resp.UpdateOption(dhcpv4.OptTFTPServerName(s.ServerAddress.String()))

		token, err := model.NewFirmwareToken(req.ClientHWAddr.String(), fwtype)
		if err != nil {
			return fmt.Errorf("iPXE firmware - failed to generated signed Firmware token")
		}
		resp.UpdateOption(dhcpv4.OptBootFileName(fmt.Sprintf("tftp://%s/%s", s.ServerAddress, token)))

	case firmware.EFI386, firmware.EFI64:
		// In theory, the response we send for FirmwareX86PC should
		// also work for EFI. However, some UEFI firmwares don't
		// support PXE properly, and will ignore ProxyDHCP responses
		// that try to bypass boot server discovery control.
		//
		// On the other hand, seemingly all firmwares support a
		// variant of the protocol where option 43 is not
		// provided. They behave as if option 43 had pointed them to a
		// PXE boot server on port 4011 of the machine sending the
		// ProxyDHCP response. Looking at TianoCore sources, I believe
		// this is the BINL protocol, which is Microsoft-specific and
		// lacks a specification. However, empirically, this code
		// seems to work.
		//
		// So, for EFI, we just provide a server name and filename,
		// and expect to be called again on port 4011 (which is in
		// pxe.go).
		log.Printf("EFI boot PXE client")
		resp.UpdateOption(dhcpv4.OptTFTPServerName(s.ServerAddress.String()))

		token, err := model.NewFirmwareToken(req.ClientHWAddr.String(), fwtype)
		if err != nil {
			return fmt.Errorf("EFI failed to generated signed Firmware token")
		}
		resp.UpdateOption(dhcpv4.OptBootFileName(token))

	case firmware.GRENDEL:
		// We've already gone through one round of chainloading, now
		// we can finally chainload to HTTP for the actual boot
		// script.
		host := s.Hostname
		if host == "" {
			host = s.ServerAddress.String()
		}

		token, err := model.NewBootToken(req.ClientHWAddr.String(), "default", fwtype)
		if err != nil {
			return fmt.Errorf("GRENDEL failed to generated signed Boot token")
		}
		ipxeUrl := fmt.Sprintf("%s://%s:%d/_/ipxe?token=%s", s.HTTPScheme, host, s.HTTPPort, token)
		log.Printf("Sending URL to iPXE script: %s", ipxeUrl)
		resp.UpdateOption(dhcpv4.OptBootFileName(ipxeUrl))

	default:
		return fmt.Errorf("unknown firmware type %d", fwtype)
	}

	return nil
}
