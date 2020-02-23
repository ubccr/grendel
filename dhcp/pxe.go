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
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/util"
)

const (
	DefaultPXEPort = 4011
)

type PXEServer struct {
	DB            model.DataStore
	ListenAddress net.IP
	ServerAddress net.IP
	Port          int
	srv           *server4.Server
	log           *logrus.Entry
	conn          net.PacketConn
	quit          chan interface{}
	wg            sync.WaitGroup
}

func NewPXEServer(db model.DataStore, address string) (*PXEServer, error) {
	s := &PXEServer{DB: db, log: logger.GetLogger("PXE"), quit: make(chan interface{})}

	if address == "" {
		address = fmt.Sprintf("%s:%d", net.IPv4zero.String(), DefaultPXEPort)
	}

	ipStr, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	s.Port = port

	ip := net.ParseIP(ipStr)
	if ip == nil || ip.To4() == nil {
		return nil, fmt.Errorf("Invalid IPv4 address: %s", ipStr)
	}

	s.ListenAddress = ip

	if !ip.To4().Equal(net.IPv4zero) {
		s.ServerAddress = ip
		return s, nil
	}

	ipaddr, err := util.GetFirstExternalIPFromInterfaces()
	if err != nil {
		return nil, err
	}

	s.ServerAddress = ipaddr

	return s, nil
}

func (s *PXEServer) pxeHandler4(peer net.Addr, req *dhcpv4.DHCPv4) {
	host, err := s.DB.LoadHostFromMAC(req.ClientHWAddr.String())
	if err != nil {
		if !errors.Is(err, model.ErrNotFound) {
			s.log.Errorf("failed to find host: %s", err)
		}
		return
	}

	if !host.Provision {
		s.log.Infof("Host %s not set to provision: %s", host.Name, req.ClientHWAddr.String())
		return
	}

	s.log.Debugf("Received DHCPv4 packet")
	s.log.Debugf(req.Summary())

	if req.OpCode != dhcpv4.OpcodeBootRequest {
		s.log.Warningf("not a BootRequest, ignoring")
		return
	}

	if !req.Options.Has(dhcpv4.OptionClientSystemArchitectureType) {
		s.log.Infof("ignoring packet - missing client system architecture type")
		return
	}

	fwtype, err := firmware.DetectBuild(req.ClientArch(), "")
	if err != nil {
		s.log.Errorf("failed to get firmware: %s", err)
		return
	}
	if host.Firmware != 0 {
		s.log.Infof("Overriding firmware for host: %s", req.ClientHWAddr.String())
		fwtype = host.Firmware
	}

	s.log.Infof("Received valid request %s - %d", req.ClientHWAddr, fwtype)

	resp, err := dhcpv4.NewReplyFromRequest(req,
		dhcpv4.WithBroadcast(false),
		dhcpv4.WithServerIP(s.ServerAddress),
		dhcpv4.WithClientIP(req.ClientIPAddr),
		dhcpv4.WithMessageType(dhcpv4.MessageTypeAck),
		dhcpv4.WithOption(dhcpv4.OptClassIdentifier("PXEClient")),
		dhcpv4.WithOption(dhcpv4.OptServerIdentifier(s.ServerAddress)),
	)
	if err != nil {
		s.log.Errorf("failed to build reply: %v", err)
		return
	}

	if req.Options.Has(dhcpv4.OptionClientMachineIdentifier) {
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionClientMachineIdentifier, req.Options.Get(dhcpv4.OptionClientMachineIdentifier)))
	}

	token, err := model.NewFirmwareToken(req.ClientHWAddr.String(), fwtype)
	if err != nil {
		s.log.Errorf("Failed to generated signed firmware token: %v", err)
		return
	}
	resp.BootFileName = token

	s.log.Debugf("Sending response")
	s.log.Debugf(resp.Summary())

	if _, err := s.conn.WriteTo(resp.ToBytes(), peer); err != nil {
		s.log.Errorf("UDP write to %v failed: %v", peer, err)
	}
}

func (s *PXEServer) Serve() error {
	listener := &net.UDPAddr{
		IP:   s.ListenAddress,
		Port: s.Port,
	}

	intf := ""
	if !s.ListenAddress.To4().Equal(net.IPv4zero) {
		var err error
		intf, _, err = util.GetInterfaceFromIP(s.ListenAddress)
		if err != nil {
			return err
		}

		listener = &net.UDPAddr{Port: s.Port}
		log.Printf("Binding to interface: %s", intf)
	}

	conn, err := server4.NewIPv4UDPConn(intf, listener)
	if err != nil {
		return err
	}

	s.conn = conn
	s.log.Infof("Server listening on: %s:%d", s.ListenAddress, s.Port)
	return s.serve()
}

func (s *PXEServer) serve() error {
	var buf [1500]byte
	for {
		n, peer, err := s.conn.ReadFrom(buf[:])
		if err != nil {
			select {
			case <-s.quit:
				return nil
			default:
				log.Errorf("Failed to read packet: %s", err)
			}
		} else {
			log.Printf("Handling request from %v", peer)

			m, err := dhcpv4.FromBytes(buf[:n])
			if err != nil {
				log.Printf("Error parsing DHCPv4 request: %v", err)
				continue
			}

			upeer, ok := peer.(*net.UDPAddr)
			if !ok {
				log.Printf("Not a UDP connection? Peer is %s", peer)
				continue
			}
			// Set peer to broadcast if the client did not have an IP.
			if upeer.IP == nil || upeer.IP.To4().Equal(net.IPv4zero) {
				upeer = &net.UDPAddr{
					IP:   net.IPv4bcast,
					Port: upeer.Port,
				}
			}

			s.wg.Add(1)
			go func() {
				s.pxeHandler4(upeer, m)
				s.wg.Done()
			}()
		}
	}

	return nil
}

func (s *PXEServer) Shutdown(ctx context.Context) error {
	close(s.quit)
	if s.conn == nil {
		return nil
	}

	defer s.conn.Close()
	s.conn.SetReadDeadline(CancelTime)

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return nil
		}
	}
}
