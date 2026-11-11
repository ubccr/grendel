// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bougou/go-ipmi/pkg/bmc"
	"github.com/bougou/go-ipmi/pkg/clock"
	"github.com/bougou/go-ipmi/pkg/hal/mock"
	"github.com/bougou/go-ipmi/pkg/server"
	"github.com/bougou/go-ipmi/pkg/transport/udp"
)

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// consoleFixture runs go-ipmi's simulated BMC on a loopback UDP port with a fake serial port behind it
func consoleFixture(t *testing.T) (*Service, Target, *mock.FakeConsoleConn) {
	t.Helper()
	fake := &mock.FakeConsoleConn{}
	m := mock.New()
	m.SetConsole(&mock.Console{Conn: fake})

	b := bmc.New(bmc.DeviceInfo{IPMIVersion: 0x20}, [16]byte{}, m, bmc.WithClock(clock.Real))
	user, err := b.Users.Add(2, "root")
	if err != nil {
		t.Fatal(err)
	}
	user.SetPassword([]byte("calvin"))
	user.Enabled = true
	user.ChannelAccess[1] = bmc.UserChannelAccess{MaxPrivilege: bmc.PrivilegeLevelAdministrator, Enabled: true}

	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	srv := server.NewServer(b, udp.Wrap(pc))
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = srv.Serve(ctx) }()
	t.Cleanup(func() {
		cancel()
		_ = srv.Close()
		_ = pc.Close()
	})

	svc := New(Config{
		Credentials: []Credential{{User: "nobody", Password: "wrong"}, {User: "root", Password: "calvin"}},
		DialTimeout: 2 * time.Second,
		IPMIPort:    pc.LocalAddr().(*net.UDPAddr).Port,
	})
	return svc, Target{Name: "cpn-f01-04", Addr: "127.0.0.1"}, fake
}

func waitFor(t *testing.T, what string, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

type session struct {
	in     *io.PipeWriter
	out    *syncBuffer
	cancel context.CancelFunc
	done   chan error
	active chan struct{}
}

func openConsole(svc *Service, target Target, force bool) *session {
	inR, inW := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	s := &session{in: inW, out: &syncBuffer{}, cancel: cancel, done: make(chan error, 1), active: make(chan struct{})}
	go func() {
		s.done <- svc.Console(ctx, target, inR, s.out, ConsoleOptions{Force: force, OnActive: func() { close(s.active) }})
	}()
	return s
}

func (s *session) wait(t *testing.T) error {
	t.Helper()
	select {
	case err := <-s.done:
		return err
	case <-time.After(10 * time.Second):
		t.Fatal("console did not return")
		return nil
	}
}

func TestConsoleRoundTrip(t *testing.T) {
	svc, target, fake := consoleFixture(t)
	s := openConsole(svc, target, false)
	defer s.cancel()

	select {
	case <-s.active:
	case err := <-s.done:
		t.Fatalf("console ended before attaching: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("console never attached")
	}

	fake.FeedRX([]byte("cpn-f01-04 login: "))
	waitFor(t, "console output", func() bool { return strings.Contains(s.out.String(), "login: ") })

	if _, err := s.in.Write([]byte("root\r")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "keystrokes on the serial port", func() bool { return fake.TXString() == "root\r" })

	s.cancel()
	if err := s.wait(t); err != nil {
		t.Fatalf("a viewer leaving should end the console cleanly, got %v", err)
	}
}

func TestConsoleBusyAndForce(t *testing.T) {
	svc, target, fake := consoleFixture(t)
	first := openConsole(svc, target, false)
	defer first.cancel()
	<-first.active

	second := openConsole(svc, target, false)
	if err := second.wait(t); !errors.Is(err, ErrConsoleBusy) {
		t.Fatalf("a second console without force should be busy, got %v", err)
	}

	third := openConsole(svc, target, true)
	defer third.cancel()
	select {
	case <-third.active:
	case err := <-third.done:
		t.Fatalf("force should take the console over, got %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("forced console never attached")
	}

	if err := first.wait(t); err == nil || !strings.Contains(err.Error(), "taken over") {
		t.Fatalf("the session taken over should say so, got %v", err)
	}

	// the first session cleaning up must not end the one that took over
	fake.FeedRX([]byte("still here"))
	waitFor(t, "output on the forced console", func() bool { return strings.Contains(third.out.String(), "still here") })
}

func TestConsoleNoCredentialAccepted(t *testing.T) {
	svc, target, _ := consoleFixture(t)
	svc.cfg.Credentials = []Credential{{User: "root", Password: "wrong"}}

	s := openConsole(svc, target, false)
	err := s.wait(t)
	if err == nil || !strings.Contains(err.Error(), "no configured credentials were accepted") {
		t.Fatalf("expected the rejected login to be reported, got %v", err)
	}
}
