// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package ping

import (
	"context"
	"net/netip"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	loopback := netip.MustParseAddr("127.0.0.1")
	// TEST-NET-1 is reserved for documentation, so nothing there answers
	nobody := netip.MustParseAddr("192.0.2.1")

	answered, err := Ping(context.Background(), []netip.Addr{loopback, nobody}, 2, 300*time.Millisecond)
	if err != nil {
		t.Skipf("can't open an ICMP socket here: %v", err)
	}
	if !answered[loopback] {
		t.Errorf("%s didn't answer", loopback)
	}
	if answered[nobody] {
		t.Errorf("%s answered", nobody)
	}
}

func TestPingNothing(t *testing.T) {
	answered, err := Ping(context.Background(), nil, 2, time.Second)
	if err != nil || len(answered) != 0 {
		t.Fatalf("got %v, %v; want no answers", answered, err)
	}
}

func TestPingFailsWhenNothingCanBeSent(t *testing.T) {
	if _, err := Ping(context.Background(), []netip.Addr{netip.MustParseAddr("127.0.0.1")}, 1, 100*time.Millisecond); err != nil {
		t.Skipf("can't open an ICMP socket here: %v", err)
	}

	// the kernel refuses a send to the broadcast address on a socket not allowed to broadcast, standing in for a firewall refusing every request
	answered, err := Ping(context.Background(), []netip.Addr{netip.MustParseAddr("255.255.255.255")}, 2, 100*time.Millisecond)
	if err == nil {
		t.Fatalf("got %v and no error, want an error since nothing was sent", answered)
	}
}
