package dns

import (
	"net"
	"strings"
	"sync"

	"github.com/miekg/dns"
	"github.com/ubccr/grendel/client"
)

// This code is based of the hosts plugin from coredns
// https://github.com/coredns/coredns/tree/master/plugin/hosts

type handler struct {
	client *client.Client
	name4  map[string][]net.IP
	name6  map[string][]net.IP
	addr   map[string][]string
	ttl    uint32

	sync.RWMutex
}

func NewHandler(client *client.Client, ttl uint32) (*handler, error) {
	h := &handler{
		client: client,
		ttl:    ttl,
		name4:  make(map[string][]net.IP),
		name6:  make(map[string][]net.IP),
		addr:   make(map[string][]string),
	}

	hostList, err := h.client.HostList()
	if err != nil {
		return nil, err
	}

	for _, host := range hostList {
		for _, nic := range host.Interfaces {
			name := Normalize(nic.FQDN)
			family := 0
			if nic.IP.To4() != nil {
				family = 1
			} else {
				family = 2
			}
			switch family {
			case 1:
				h.name4[name] = append(h.name4[name], nic.IP)
			case 2:
				h.name6[name] = append(h.name6[name], nic.IP)
			default:
				continue
			}
			h.addr[nic.IP.String()] = append(h.addr[nic.IP.String()], name)
		}
	}

	return h, nil
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	qname := h.Name(r)
	answers := []dns.RR{}
	switch h.QType(r) {
	case dns.TypePTR:
		names := h.LookupStaticAddr(ExtractAddressFromReverse(qname))
		answers = h.ptr(qname, h.ttl, names)
	case dns.TypeAAAA:
		ips := h.LookupStaticHostV6(qname)
		answers = aaaa(qname, h.ttl, ips)
	case dns.TypeA:
		ips := h.LookupStaticHostV4(qname)
		answers = a(qname, h.ttl, ips)
	}

	if len(answers) != 0 {
		m.Authoritative = true
		m.Answer = answers
		m.SetRcode(r, dns.RcodeSuccess)
	} else {
		// XXX consider sending back NXDOMAIN here
		m.SetRcode(r, dns.RcodeNameError)
	}

	w.WriteMsg(m)
}

func (h *handler) LookupStaticHostV4(host string) []net.IP {
	host = strings.ToLower(host)
	return h.lookupStaticHost(h.name4, host)
}

func (h *handler) LookupStaticHostV6(host string) []net.IP {
	host = strings.ToLower(host)
	return h.lookupStaticHost(h.name6, host)
}

func (h *handler) LookupStaticAddr(addr string) []string {
	addr = parseIP(addr).String()
	if addr == "" {
		return nil
	}

	h.RLock()
	defer h.RUnlock()

	hosts := h.addr[addr]

	if len(hosts) == 0 {
		return nil
	}

	hostsCp := make([]string, len(hosts))
	copy(hostsCp, hosts)
	return hostsCp
}

// parseIP calls discards any v6 zone info, before calling net.ParseIP.
func parseIP(addr string) net.IP {
	if i := strings.Index(addr, "%"); i >= 0 {
		// discard ipv6 zone
		addr = addr[0:i]
	}

	return net.ParseIP(addr)
}

// a takes a slice of net.IPs and returns a slice of A RRs.
func a(zone string, ttl uint32, ips []net.IP) []dns.RR {
	answers := make([]dns.RR, len(ips))
	for i, ip := range ips {
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl}
		r.A = ip
		answers[i] = r
	}
	return answers
}

// aaaa takes a slice of net.IPs and returns a slice of AAAA RRs.
func aaaa(zone string, ttl uint32, ips []net.IP) []dns.RR {
	answers := make([]dns.RR, len(ips))
	for i, ip := range ips {
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: ttl}
		r.AAAA = ip
		answers[i] = r
	}
	return answers
}

// ptr takes a slice of host names and filters out the ones that aren't in Origins, if specified, and returns a slice of PTR RRs.
func (h *handler) ptr(zone string, ttl uint32, names []string) []dns.RR {
	answers := make([]dns.RR, len(names))
	for i, n := range names {
		r := new(dns.PTR)
		r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: ttl}
		r.Ptr = dns.Fqdn(n)
		answers[i] = r
	}
	return answers
}

func (h *handler) lookupStaticHost(m map[string][]net.IP, host string) []net.IP {
	h.RLock()
	defer h.RUnlock()

	if len(m) == 0 {
		return nil
	}

	ips, ok := m[host]
	if !ok {
		return nil
	}
	ipsCp := make([]net.IP, len(ips))
	copy(ipsCp, ips)
	return ipsCp
}

func (h *handler) Name(r *dns.Msg) string {
	if len(r.Question) == 0 {
		return "."
	}

	return strings.ToLower(dns.Name(r.Question[0].Name).String())
}

func (h *handler) QType(r *dns.Msg) uint16 {
	if len(r.Question) == 0 {
		return 0
	}

	return r.Question[0].Qtype
}
