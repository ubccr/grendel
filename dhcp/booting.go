package dhcp

import (
	"fmt"

	"github.com/insomniacslk/dhcp/dhcpv4"
	log "github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/model"
)

func (s *Server) bootingHandler4(req, resp *dhcpv4.DHCPv4) error {
	fwt, err := dhcpv4.GetUint16(dhcpv4.OptionClientSystemArchitectureType, req.Options)
	if err != nil {
		return fmt.Errorf("missing DHCP option 93 system architecture")
	}

	userClass := ""
	if req.Options.Has(dhcpv4.OptionUserClassInformation) {
		userClass = string(req.Options.Get(dhcpv4.OptionUserClassInformation))
	}

	fwtype, err := model.NewFirmwareFromDHCP(fwt, userClass)
	if err != nil {
		return fmt.Errorf("Failed to get PXE firmware from DHCP: %s", err)
	}

	mach, err := model.NewMachineFromDHCP(req.ClientHWAddr, fwt)
	if err != nil {
		return fmt.Errorf("Failed to get machine from DHCP: %s", err)
	}

	guid := req.Options.Get(dhcpv4.OptionClientMachineIdentifier)
	switch len(guid) {
	case 0:
		// A missing GUID is invalid according to the spec, however
		// there are PXE ROMs in the wild that omit the GUID and still
		// expect to boot. The only thing we do with the GUID is
		// mirror it back to the client if it's there, so we might as
		// well accept these buggy ROMs.
	case 17:
		if guid[0] != 0 {
			return fmt.Errorf("malformed client GUID (option 97), leading byte must be zero")
		}
	default:
		return fmt.Errorf("malformed client GUID (option 97), wrong size")
	}

	log.Infof("Got valid request to boot %s (%s) %d", mach.MAC, mach.Arch, fwtype)

	switch fwtype {
	case model.FirmwareX86PC:
		// This is completely standard PXE: we tell the PXE client to
		// bypass all the boot discovery rubbish that PXE supports,
		// and just load a file from TFTP.
		log.Printf("DHCP - FirmwareX86PC telling PXE client to bypass all boot discovery")

		pxe := dhcpv4.OptionsFromList(dhcpv4.OptGeneric(dhcpv4.GenericOptionCode(6), []byte{8}))
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionVendorSpecificInformation, pxe.ToBytes()))

		resp.UpdateOption(dhcpv4.OptTFTPServerName(s.ServerAddress.String()))
		resp.UpdateOption(dhcpv4.OptBootFileName(fmt.Sprintf("%s/%d", mach.MAC, fwtype)))

	case model.FirmwareX86Ipxe:
		log.Printf("DHCP - FirmwareX86Ipxe telling PXE client to boot tftp")
		// Almost standard PXE, but the boot filename needs to be a URL.

		// PXE Boot Server Discovery Control - bypass, just boot from filename.
		pxe := dhcpv4.OptionsFromList(dhcpv4.OptGeneric(dhcpv4.GenericOptionCode(6), []byte{8}))
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionVendorSpecificInformation, pxe.ToBytes()))

		resp.UpdateOption(dhcpv4.OptTFTPServerName(s.ServerAddress.String()))
		resp.UpdateOption(dhcpv4.OptBootFileName(fmt.Sprintf("tftp://%s/%s/%d", s.ServerAddress, mach.MAC, fwtype)))

	case model.FirmwareEFI32, model.FirmwareEFI64, model.FirmwareEFIBC:
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
		log.Printf("DHCP - EFI boot PXE client")
		resp.UpdateOption(dhcpv4.OptTFTPServerName(s.ServerAddress.String()))
		resp.UpdateOption(dhcpv4.OptBootFileName(fmt.Sprintf("%s/%d", mach.MAC, fwtype)))

	case model.FirmwarePixiecoreIpxe:
		// We've already gone through one round of chainloading, now
		// we can finally chainload to HTTP for the actual boot
		// script.
		host := s.Hostname
		if host == "" {
			host = s.ServerAddress.String()
		}
		ipxeUrl := fmt.Sprintf("%s://%s:%d/_/ipxe?arch=%d&mac=%s", s.HTTPScheme, host, s.HTTPPort, mach.Arch, mach.MAC)
		log.Printf("DHCP - FirmwarePixiecoreIpxe sending URL to iPXE script: %s", ipxeUrl)
		resp.UpdateOption(dhcpv4.OptBootFileName(ipxeUrl))

	default:
		return fmt.Errorf("unknown firmware type %d", fwtype)
	}

	return nil
}
