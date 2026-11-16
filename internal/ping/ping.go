// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package ping checks whether IPv4 addresses answer an ICMP echo request
package ping

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// protocolICMP is the IANA protocol number of ICMP, which icmp.ParseMessage takes
const protocolICMP = 1

// Ping sends an echo request to each address and reports the ones that answered. An address is asked again, tries times in all, until it answers, with wait between tries. When no request can be sent at all, as when a firewall refuses them, it fails rather than report that nothing answered. Echo goes out on an unprivileged ICMP socket where net.ipv4.ping_group_range allows one, otherwise on a raw socket, which needs CAP_NET_RAW
func Ping(ctx context.Context, addrs []netip.Addr, tries int, wait time.Duration) (map[netip.Addr]bool, error) {
	answered := make(map[netip.Addr]bool, len(addrs))
	if len(addrs) == 0 {
		return answered, nil
	}

	privileged := false
	conn, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		privileged = true
		conn, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			return nil, err
		}
	}
	defer conn.Close()

	// on an unprivileged socket the kernel sets the ID itself and only hands this socket its own replies, so the ID is only checked on a raw socket, which sees every reply on the host
	id := os.Getpid() & 0xffff
	buf := make([]byte, 1500)
	wanted := make(map[netip.Addr]bool, len(addrs))
	for _, a := range addrs {
		wanted[a] = a.Is4()
	}

	for try := range tries {
		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Body: &icmp.Echo{ID: id, Seq: try, Data: []byte("grendel")},
		}
		b, err := msg.Marshal(nil)
		if err != nil {
			return nil, err
		}

		waiting := 0
		var sendErr error
		for addr, want := range wanted {
			if !want || answered[addr] {
				continue
			}
			var dst net.Addr = &net.IPAddr{IP: addr.AsSlice()}
			if !privileged {
				dst = &net.UDPAddr{IP: addr.AsSlice()}
			}
			// one address that can't be sent to, like one with no route, simply doesn't answer
			if _, err := conn.WriteTo(b, dst); err != nil {
				sendErr = err
				continue
			}
			waiting++
		}
		if waiting == 0 {
			if try == 0 && sendErr != nil {
				return nil, fmt.Errorf("sending echo requests: %w", sendErr)
			}
			break
		}

		deadline := time.Now().Add(wait)
		if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
			deadline = d
		}
		if err := conn.SetReadDeadline(deadline); err != nil {
			return nil, err
		}
		for waiting > 0 {
			n, peer, err := conn.ReadFrom(buf)
			if err != nil {
				var ne net.Error
				if errors.As(err, &ne) && ne.Timeout() {
					break
				}
				return nil, err
			}
			msg, err := icmp.ParseMessage(protocolICMP, buf[:n])
			if err != nil || msg.Type != ipv4.ICMPTypeEchoReply {
				continue
			}
			echo, ok := msg.Body.(*icmp.Echo)
			if !ok || (privileged && echo.ID != id) {
				continue
			}
			from, ok := peerAddr(peer)
			if !ok || answered[from] || !wanted[from] {
				continue
			}
			answered[from] = true
			waiting--
		}
		if err := ctx.Err(); err != nil {
			return answered, err
		}
	}

	return answered, nil
}

func peerAddr(peer net.Addr) (netip.Addr, bool) {
	var ip net.IP
	switch p := peer.(type) {
	case *net.UDPAddr:
		ip = p.IP
	case *net.IPAddr:
		ip = p.IP
	default:
		return netip.Addr{}, false
	}
	addr, ok := netip.AddrFromSlice(ip)
	return addr.Unmap(), ok
}
