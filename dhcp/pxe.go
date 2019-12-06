package dhcp

import (
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/model"
)

func (s *Server) pxeHandler4(conn net.PacketConn, peer net.Addr, req *dhcpv4.DHCPv4) {
	log.Debugf("PXEServer Received DHCPv4 packet")
	log.Debugf(req.Summary())

	if req.OpCode != dhcpv4.OpcodeBootRequest {
		log.Warningf("PXEServer not a BootRequest, ignoring")
		return
	}

	if !req.Options.Has(dhcpv4.OptionClientSystemArchitectureType) {
		log.Infof("PXEServer ignoring packet - missing client system architecture type")
		return
	}

	fwtype, err := firmware.DetectBuild(req.ClientArch(), "")
	if err != nil {
		log.Errorf("PXEServer failed to get firmware: %s", err)
		return
	}

	log.Infof("PXEServer received valid request %s - %d", req.ClientHWAddr, fwtype)

	resp, err := dhcpv4.NewReplyFromRequest(req,
		dhcpv4.WithBroadcast(false),
		dhcpv4.WithServerIP(s.ServerAddress),
		dhcpv4.WithClientIP(req.ClientIPAddr),
		dhcpv4.WithMessageType(dhcpv4.MessageTypeAck),
		dhcpv4.WithOption(dhcpv4.OptClassIdentifier("PXEClient")),
		dhcpv4.WithOption(dhcpv4.OptServerIdentifier(s.ServerAddress)),
	)
	if err != nil {
		log.Printf("PXEServer failed to build reply: %v", err)
		return
	}

	if req.Options.Has(dhcpv4.OptionClientMachineIdentifier) {
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionClientMachineIdentifier, req.Options.Get(dhcpv4.OptionClientMachineIdentifier)))
	}

	token, err := model.NewFirmwareToken(req.ClientHWAddr.String(), fwtype)
	if err != nil {
		log.Errorf("Failed to generated signed PXE token")
		return
	}
	resp.BootFileName = token

	log.Debugf("PXEServer sending response")
	log.Debugf(resp.Summary())

	if _, err := conn.WriteTo(resp.ToBytes(), peer); err != nil {
		log.Printf("PXEServer conn.Write to %v failed: %v", peer, err)
	}
}
