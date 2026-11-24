// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/internal/dns"
	"github.com/ubccr/grendel/internal/settings"
)

func init() {

	register(&Service{
		Name: "dns",
		New:  newDNS,
	})
}

func newDNS() (Runner, error) {
	srv, err := dns.NewServer(DB, getListenAddress("dns.listen"), settings.DNSTTL.Get())
	if err != nil {
		return nil, err
	}

	if fwAddr := settings.DNSForward.Get(); fwAddr != "" {
		shared.Log.Debugf("dns.forward address set, using: %s", fwAddr)
	}

	return srv, nil
}
