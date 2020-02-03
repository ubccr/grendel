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
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"golang.org/x/net/bpf"
	"golang.org/x/net/ipv4"
)

type Snooper struct {
	Port    int
	Handler func(req *dhcpv4.DHCPv4)
	conn    *ipv4.RawConn
	quit    chan interface{}
	wg      sync.WaitGroup
}

func NewSnooper(address string, handler func(req *dhcpv4.DHCPv4)) (*Snooper, error) {
	if address == "" {
		address = fmt.Sprintf("%s:%d", net.IPv4zero.String(), dhcpv4.ServerPort)
	}

	_, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	s := &Snooper{
		Port:    port,
		Handler: handler,
		quit:    make(chan interface{}),
	}

	return s, nil
}

func (s *Snooper) Snoop() error {
	// Adopted from https://github.com/danderson/netboot/blob/master/dhcp4/conn_linux.go
	// Written by @danderson
	filter, err := bpf.Assemble([]bpf.Instruction{
		// Load IPv4 packet length
		bpf.LoadMemShift{Off: 0},
		// Get UDP dport
		bpf.LoadIndirect{Off: 2, Size: 2},
		// Correct dport?
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: uint32(s.Port), SkipFalse: 1},
		// Accept
		bpf.RetConstant{Val: 1500},
		// Ignore
		bpf.RetConstant{Val: 0},
	})
	if err != nil {
		return err
	}

	conn, err := net.ListenPacket("ip4:17", "0.0.0.0")
	if err != nil {
		return err
	}
	defer conn.Close()

	rconn, err := ipv4.NewRawConn(conn)
	if err != nil {
		return err
	}
	if err = rconn.SetControlMessage(ipv4.FlagInterface, true); err != nil {
		return fmt.Errorf("Failed setting control message: %w", err)
	}
	if err = rconn.SetBPF(filter); err != nil {
		return fmt.Errorf("Failed setting BFP filter: %w", err)
	}

	s.conn = rconn
	return s.serve()
}

func (s *Snooper) serve() error {
	var buf [1500]byte
	for {
		_, p, _, err := s.conn.ReadFrom(buf[:])
		if err != nil {
			select {
			case <-s.quit:
				return nil
			default:
				log.Errorf("Failed to read packet: %s", err)
			}
		} else {
			if len(p) < 8 {
				log.Errorf("Invalid UDP packet too short")
				continue
			}

			m, err := dhcpv4.FromBytes(p[8:])
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
	s.conn.Close()

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
