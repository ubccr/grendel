package dns

import (
	"net"
	"strings"

	"github.com/miekg/dns"
	"github.com/ubccr/grendel/model"
)

// This code is based of the hosts plugin from coredns
// https://github.com/coredns/coredns/tree/master/plugin/hosts

type handler struct {
	db    model.Datastore
	name4 map[string][]net.IP
	name6 map[string][]net.IP
	addr  map[string][]string
	ttl   uint32
}

func NewHandler(db model.Datastore, ttl uint32) (*handler, error) {
	h := &handler{
		db:    db,
		ttl:   ttl,
		name4: make(map[string][]net.IP),
		name6: make(map[string][]net.IP),
		addr:  make(map[string][]string),
	}

	hostList, err := h.db.HostList()
	if err != nil {
		return nil, err
	}

	for _, host := range hostList {
		name := Normalize(host.FQDN)
		family := 0
		if host.IP.To4() != nil {
			family = 1
		} else {
			family = 2
		}
		switch family {
		case 1:
			h.name4[name] = append(h.name4[name], host.IP)
		case 2:
			h.name6[name] = append(h.name6[name], host.IP)
		default:
			continue
		}
		h.addr[host.IP.String()] = append(h.addr[host.IP.String()], name)
	}

	return h, nil
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	qname := m.Question[0].Name
	answers := []dns.RR{}
	switch r.Question[0].Qtype {
	case dns.TypePTR:
		names := h.LookupStaticAddr(ExtractAddressFromReverse(qname))
		if len(names) != 0 {
			answers = h.ptr(qname, h.ttl, names)
		}
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
