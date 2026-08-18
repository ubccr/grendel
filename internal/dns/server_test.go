package dns

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/internal/store/sqlstore"
	"github.com/ubccr/grendel/pkg/model"
)

var (
	serverAddr = "127.0.0.1:8053"
	clientFQDN = "test-01.example.local"
	clientIP   = netip.MustParsePrefix("10.1.0.1/24")

	// A name grendel does not know, served by the stub upstream resolver.
	upstreamFQDN = "forwarded.example.test"
	upstreamIP   = "192.0.2.53"
)

func newDNS() (*Server, error) {
	store, err := sqlstore.New(":memory:")
	if err != nil {
		return nil, err
	}

	store.StoreHost(&model.Host{
		Name: "test-01",
		Interfaces: []*model.NetInterface{
			{
				FQDN: clientFQDN,
				IP:   clientIP,
			},
		},
	})

	s, err := NewServer(store, serverAddr, 5)
	return s, err
}

func TestDns(t *testing.T) {
	assert := assert.New(t)
	s, err := newDNS()
	if err != nil {
		t.Fatal(err)
	}

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8053")
	c, err := net.ListenUDP("udp", addr)
	if err == nil {
		c.Close()
		go func() {
			err := s.Serve()
			assert.NoError(err)
			defer s.Shutdown(context.Background())
		}()
	}

	time.Sleep(time.Second * 1)

	// Check standard grendel lookup
	m1 := new(dns.Msg)
	m1.SetQuestion(clientFQDN+".", dns.TypeA)

	r1, err := dns.Exchange(m1, serverAddr)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(r1.Response)
	if len(r1.Answer) == 0 {
		t.Fatal(errors.New("r1 response is empty"))
	}
	assert.Equal(r1.Answer[0].String(), clientFQDN+".\t5\tIN\tA\t10.1.0.1")

	// Check reverse grendel lookup
	m2 := new(dns.Msg)
	m2.SetQuestion("1.0.1.10.in-addr.arpa.", dns.TypePTR)

	r2, err := dns.Exchange(m2, serverAddr)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(r2.Response)
	if len(r2.Answer) == 0 {
		t.Fatal(errors.New("r2 response is empty"))
	}
	assert.Equal(r2.Answer[0].String(), "1.0.1.10.in-addr.arpa.\t5\tIN\tPTR\ttest-01.example.local.")

	// Check non forwarded lookup
	m3 := new(dns.Msg)
	m3.SetQuestion("miekl.nl.", dns.TypeMX)

	r3, err := dns.Exchange(m3, serverAddr)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(r3.Response)
	assert.Len(r3.Answer, 0)

	upstream := newStubResolver(t, upstreamFQDN, upstreamIP)

	// viper is process global and this key leaks into every later test in the package, so put it back.
	t.Cleanup(func() { viper.Set("dns.forward", "") })
	viper.Set("dns.forward", upstream)

	// A name grendel knows is still answered locally, not forwarded.
	m4 := new(dns.Msg)
	m4.SetQuestion(clientFQDN+".", dns.TypeA)

	r4, err := dns.Exchange(m4, serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(r4.Response)
	if len(r4.Answer) == 0 {
		t.Fatal(errors.New("r4 response is empty"))
	}
	assert.Equal(r4.Answer[0].String(), clientFQDN+".\t5\tIN\tA\t10.1.0.1")

	// A name grendel does not know is forwarded upstream.
	m5 := new(dns.Msg)
	m5.SetQuestion(upstreamFQDN+".", dns.TypeA)

	r5, err := dns.Exchange(m5, serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(r5.Response)
	if len(r5.Answer) == 0 {
		t.Fatal(errors.New("r5 response is empty"))
	}
	p5 := strings.Split(r5.Answer[0].String(), "\t")
	if len(p5) != 5 {
		t.Fatal(errors.New("p5 response length is incorrect"))
	}

	assert.Equal(p5[0], upstreamFQDN+".")
	assert.Equal(p5[2], "IN")
	assert.Equal(p5[3], "A")
	assert.Equal(p5[4], upstreamIP)
}

// newStubResolver starts a throwaway DNS server on an ephemeral port that
// answers exactly one name, and returns its address.
func newStubResolver(t *testing.T, name, ip string) string {
	t.Helper()

	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	srv := &dns.Server{PacketConn: pc}
	srv.Handler = dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		if len(r.Question) > 0 && r.Question[0].Name == dns.Fqdn(name) {
			rr, err := dns.NewRR(dns.Fqdn(name) + "\t5\tIN\tA\t" + ip)
			if err == nil {
				m.Answer = append(m.Answer, rr)
			}
		}
		w.WriteMsg(m)
	})

	started := make(chan struct{})
	srv.NotifyStartedFunc = func() { close(started) }
	go srv.ActivateAndServe()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("stub resolver did not start")
	}
	t.Cleanup(func() { srv.Shutdown() })

	return pc.LocalAddr().String()
}
