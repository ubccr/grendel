package dns

import (
	"context"
	"errors"
	"net/netip"
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

	go func() {
		err := s.Serve()
		assert.NoError(err)
		defer s.Shutdown(context.Background())
	}()

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

	viper.Set("dns.forward", "1.1.1.1:53")

	// Check grendel lookup with forward address set
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

	// Check forwarded MX lookup
	m5 := new(dns.Msg)
	m5.SetQuestion("miek.nl.", dns.TypeMX)

	r5, err := dns.Exchange(m5, serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(r5.Response)
	if len(r5.Answer) == 0 {
		t.Fatal(errors.New("r5 response is empty"))
	}
	assert.Equal(r5.Answer[0].String(), "miek.nl.\t21600\tIN\tMX\t1 aspmx.l.google.com.")
}
