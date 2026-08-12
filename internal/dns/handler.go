// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package dns

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/util"
)

type handler struct {
	db      store.Store
	ttl     uint32
	timeout time.Duration
}

func NewHandler(db store.Store, ttl uint32) (*handler, error) {
	timeout := viper.GetDuration("dns.query_timeout")
	// A zero timeout would expire every context immediately.
	if timeout <= 0 {
		timeout = defaultQueryTimeout
	}

	h := &handler{
		db:      db,
		ttl:     ttl,
		timeout: timeout,
	}

	return h, nil
}

func (h *handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	qname := h.Name(r)
	answers := []dns.RR{}
	queryType := h.QType(r)

	log.WithFields(logrus.Fields{
		"query":  qname,
		"type":   dns.TypeToString[queryType],
		"client": w.RemoteAddr(),
	}).Debug("Got DNS query")

	var queryErr error
	switch queryType {
	case dns.TypePTR:
		names, err := h.db.ReverseResolve(ctx, util.ExtractAddressFromReverse(qname))
		if err != nil {
			queryErr = err
			log.WithFields(logrus.Fields{
				"qname": qname,
				"err":   err,
			}).Error("Failed to reverse resolve IP")
		}
		answers = h.ptr(qname, h.ttl, names)
	case dns.TypeA:
		ips, err := h.db.ResolveIPv4(ctx, qname)
		if err != nil {
			queryErr = err
			log.WithFields(logrus.Fields{
				"qname": qname,
				"err":   err,
			}).Error("Failed to resolve FQDN")
		}
		answers = a(qname, h.ttl, ips)
	}

	// Answer SERVFAIL rather than going silent. A dropped response makes the client retransmit
	if errors.Is(queryErr, context.DeadlineExceeded) {
		m.SetRcode(r, dns.RcodeServerFailure)
		if err := w.WriteMsg(m); err != nil {
			log.WithError(err).Warn("Failed to write DNS response")
		}
		return
	}

	fwAddr := viper.GetString("dns.forward")
	if len(answers) != 0 {
		m.Authoritative = true
		m.Answer = answers
		m.SetRcode(r, dns.RcodeSuccess)
	} else if len(answers) == 0 && fwAddr != "" {
		fwm, err := dns.ExchangeContext(ctx, r, fwAddr)
		if err != nil {
			log.WithFields(logrus.Fields{
				"qname": qname,
				"err":   err,
			}).Error("Failed to forward DNS")
			m.SetRcode(r, dns.RcodeServerFailure)
		} else {
			m = fwm
		}

	} else if queryType == dns.TypeAAAA || queryType == dns.TypeMX {
		// Handle returning AAAA if IPv4 record exists
		ips, err := h.db.ResolveIPv4(ctx, qname)
		if err != nil {
			log.WithFields(logrus.Fields{
				"qname": qname,
				"err":   err,
			}).Error("Failed to resolve FQDN during AAAA or MX query")
		}

		if len(ips) > 0 {
			m.SetRcode(r, dns.RcodeSuccess)
		} else {
			m.SetRcode(r, dns.RcodeNameError)
		}
	} else {
		m.SetRcode(r, dns.RcodeNameError)
	}

	if err := w.WriteMsg(m); err != nil {
		log.WithError(err).Warn("Failed to write DNS response")
	}
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
