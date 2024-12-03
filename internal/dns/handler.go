// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package dns

import (
	"net"
	"strings"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/internal/util"
)

type handler struct {
	db  model.DataStore
	ttl uint32
}

func NewHandler(db model.DataStore, ttl uint32) (*handler, error) {
	h := &handler{
		db:  db,
		ttl: ttl,
	}

	return h, nil
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	qname := h.Name(r)
	answers := []dns.RR{}

	log.Debugf("Got query %s", qname)
	switch h.QType(r) {
	case dns.TypePTR:
		names, err := h.db.ReverseResolve(util.ExtractAddressFromReverse(qname))
		if err != nil {
			log.WithFields(logrus.Fields{
				"qname": qname,
				"err":   err,
			}).Error("Failed to reverse resolve IP")
		}
		answers = h.ptr(qname, h.ttl, names)
	case dns.TypeA:
		ips, err := h.db.ResolveIPv4(qname)
		if err != nil {
			log.WithFields(logrus.Fields{
				"qname": qname,
				"err":   err,
			}).Error("Failed to resolve FQDN")
		}
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

// The code below was adopted from the hosts plugin from coredns
// https://github.com/coredns/coredns/tree/master/plugin/hosts
// Copyright coredns authors Apache License

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
