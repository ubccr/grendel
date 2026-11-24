// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import "time"

var (
	DNSListen = register[string](Option{
		Key:      "dns.listen",
		Section:  "dns",
		Summary:  "Listen address",
		Help:     "Address and UDP port the DNS server binds to. Use 0.0.0.0 to answer on every interface.",
		Kind:     KindString,
		Validate: validAddr,
		Default:  "0.0.0.0:53",
		Required: true,
		Apply:    ApplyRestart,
		Flag:     "dns-listen",
	})
	DNSTTL = register[int](Option{
		Key:     "dns.ttl",
		Section: "dns",
		Summary: "Record TTL",
		Help:    "How long resolvers may cache an answer from grendel. Lower it while hosts are being renumbered, raise it to cut query load.",
		Kind:    KindInt,
		Default: 300,
		Unit:    "seconds",
		Min:     ptr(0),
		Apply:   ApplyRestart,
		Flag:    "dns-ttl",
	})
	DNSForward = register[string](Option{
		Key:      "dns.forward",
		Section:  "dns",
		Summary:  "Upstream resolver",
		Help:     "Address and port of a resolver that gets the queries grendel has no answer for, such as 1.1.1.1:53. Leave empty to answer only for known hosts.",
		Kind:     KindString,
		Validate: validAddr,
		Default:  "",
		Apply:    ApplyLive,
		Flag:     "dns-forward",
	})
	DNSQueryTimeout = register[time.Duration](Option{
		Key:     "dns.query_timeout",
		Section: "dns",
		Summary: "Query timeout",
		Help:    "How long a single query may take before grendel answers SERVFAIL. Most resolvers give up after 2s, so a longer value only holds database connections for clients that have already left.",
		Kind:    KindDuration,
		Default: 2 * time.Second,
		Apply:   ApplyRestart,
	})
)
