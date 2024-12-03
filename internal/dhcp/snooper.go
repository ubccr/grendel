// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package dhcp

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/ubccr/grendel/internal/util"
)

type Snooper struct {
	ListenAddress net.IP
	Port          int
	Handler       func(req *dhcpv4.DHCPv4)
	conn          net.PacketConn
	quit          chan interface{}
	wg            sync.WaitGroup
}

func NewSnooper(address string, handler func(req *dhcpv4.DHCPv4)) (*Snooper, error) {
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

	ip := net.ParseIP(ipStr)
	if ip == nil || ip.To4() == nil {
		return nil, fmt.Errorf("Invalid IPv4 address: %s", ipStr)
	}

	s := &Snooper{
		Port:          port,
		ListenAddress: ip,
		Handler:       handler,
		quit:          make(chan interface{}),
	}

	return s, nil
}

func (s *Snooper) Snoop() error {
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
	return s.serve()
}

func (s *Snooper) serve() error {
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
			log.Printf("Snooping request from %v", peer)

			m, err := dhcpv4.FromBytes(buf[:n])
			if err != nil {
				log.Errorf("Error parsing DHCPv4 request: %v", err)
				continue
			}

			s.wg.Add(1)
			go func() {
				s.Handler(m)
				s.wg.Done()
			}()
		}
	}

	return nil
}

func (s *Snooper) Shutdown(ctx context.Context) error {
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

	return nil
}
