// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package dhcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/util"
	"golang.org/x/net/ipv4"
)

var log = logger.GetLogger("DHCP")

type Server struct {
	ListenAddress  net.IP
	ServerAddress  net.IP
	InterfaceIPMap map[int]net.IP
	Port           int
	ProxyOnly      bool
	DB             store.Store
	LeaseTime      time.Duration
	conn           *ipv4.PacketConn
	quit           chan interface{}
	wg             sync.WaitGroup
}

func NewServer(db store.Store, address string) (*Server, error) {
	s := &Server{DB: db, quit: make(chan interface{})}

	if address == "" {
		address = fmt.Sprintf("%s:%d", net.IPv4zero.String(), dhcpv4.ServerPort)
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

	log.Infof("Using default ServerAddress: %s", ipaddr)
	s.ServerAddress = ipaddr

	intfMap, err := util.GetInterfaceIPMap()
	if err != nil {
		return nil, err
	}

	s.InterfaceIPMap = intfMap

	return s, nil
}

func (s *Server) mainHandler4(peer *net.UDPAddr, req *dhcpv4.DHCPv4, oob *ipv4.ControlMessage) {
	if req.OpCode != dhcpv4.OpcodeBootRequest {
		log.Debugf("Ignoring not a BootRequest")
		return
	}

	host, err := s.DB.LoadHostFromMAC(req.ClientHWAddr.String())
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			log.Debugf("Ignoring unknown client mac address: %s", req.ClientHWAddr)
		} else {
			log.Errorf("Failed to find host from database: %s", err)
		}
		return
	}

	serverIP := s.ServerAddress
	// Use the IP address of the interface the request came in on for the
	// ServerIP if available.
	if intfIP, ok := s.InterfaceIPMap[oob.IfIndex]; ok {
		serverIP = intfIP
	}

	resp, err := dhcpv4.NewReplyFromRequest(req,
		dhcpv4.WithServerIP(serverIP),
		dhcpv4.WithMessageType(dhcpv4.MessageTypeOffer),
		dhcpv4.WithOption(dhcpv4.OptClassIdentifier("PXEClient")),
		dhcpv4.WithOption(dhcpv4.OptServerIdentifier(serverIP)),
	)
	if err != nil {
		log.Printf("DHCP failed to build reply: %v", err)
		return
	}

	// Copy hop count? is this needed?
	resp.HopCount = req.HopCount

	if req.Options.Has(dhcpv4.OptionClientMachineIdentifier) {
		resp.UpdateOption(dhcpv4.OptGeneric(dhcpv4.OptionClientMachineIdentifier, req.Options.Get(dhcpv4.OptionClientMachineIdentifier)))
	}

	switch mt := req.MessageType(); mt {
	case dhcpv4.MessageTypeDiscover:
		err := s.bootingHandler4(host, serverIP, req, resp)
		if err != nil {
			log.WithFields(logrus.Fields{
				"mac":      req.ClientHWAddr.String(),
				"host_uid": host.UID.String(),
				"err":      err,
			}).Error("Failed to add boot options to DHCP request")
			if s.ProxyOnly {
				return
			}
		}

		if !s.ProxyOnly {
			err := s.staticHandler4(host, serverIP, req, resp)
			if err != nil {
				log.Errorf("Failed to add client ip to DHCP DISCOVER: %s", err)
				return
			}
		}
	case dhcpv4.MessageTypeRequest, dhcpv4.MessageTypeInform:
		if s.ProxyOnly {
			return
		}

		err := s.staticAckHandler4(host, serverIP, req, resp)
		if err != nil {
			log.Errorf("Failed to ack DHCP REQUEST: %s", err)
			return
		}
	default:
		log.Warnf("DHCP Unhandled message type: %v", mt)
		log.Debugln(resp.Summary())
		return
	}

	peer = &net.UDPAddr{IP: net.IPv4bcast, Port: dhcpv4.ClientPort}
	if !req.GatewayIPAddr.IsUnspecified() {
		peer = &net.UDPAddr{IP: req.GatewayIPAddr, Port: dhcpv4.ServerPort}
		resp.SetBroadcast()
	} else if req.ClientIPAddr != nil && !req.ClientIPAddr.Equal(net.IPv4zero) {
		peer = &net.UDPAddr{IP: req.ClientIPAddr, Port: dhcpv4.ClientPort}
		resp.SetUnicast()
	}

	var woob *ipv4.ControlMessage
	if peer.IP.Equal(net.IPv4bcast) || peer.IP.IsLinkLocalUnicast() {
		switch {
		case oob != nil && oob.IfIndex != 0:
			woob = &ipv4.ControlMessage{IfIndex: oob.IfIndex}
		default:
			log.Errorf("mainHandler4: Did not receive interface information")
		}
	}

	log.Debugf("Sending DHCPv4 packet response")
	log.Debugln(resp.Summary())

	if _, err := s.conn.WriteTo(resp.ToBytes(), woob, peer); err != nil {
		log.Printf("DHCP write to %v failed: %v", peer, err)
	}
}

func (s *Server) Serve() error {
	listener := &net.UDPAddr{
		IP:   s.ListenAddress,
		Port: s.Port,
	}

	intf := ""
	if !s.ListenAddress.To4().Equal(net.IPv4zero) {
		iface, _, err := util.GetInterfaceFromIP(s.ListenAddress)
		if err != nil {
			return err
		}
		intf = iface
		listener = &net.UDPAddr{Port: s.Port}
		log.Printf("Binding to interface: %s", intf)
	}

	udpConn, err := server4.NewIPv4UDPConn(intf, listener)
	if err != nil {
		return err
	}

	s.conn = ipv4.NewPacketConn(udpConn)
	err = s.conn.SetControlMessage(ipv4.FlagInterface, true)
	if err != nil {
		return err
	}

	log.Infof("Server listening on: %s:%d", s.ListenAddress, s.Port)
	return s.serve()
}

func (s *Server) serve() error {
	var buf [1500]byte
	for {
		n, oob, peer, err := s.conn.ReadFrom(buf[:])
		if err != nil {
			select {
			case <-s.quit:
				return nil
			default:
				log.Errorf("Failed to read packet: %s", err)
			}
		} else {
			log.Debugf("Handling request from %v", peer)

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
				s.mainHandler4(upeer, m, oob)
				s.wg.Done()
			}()
		}
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
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
