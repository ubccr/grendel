// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/internal/config"
)

func TestEndpointsBaseURL(t *testing.T) {
	origAddr, origHostname := config.ProvisionAddr, config.ProvisionHostname
	t.Cleanup(func() {
		config.ProvisionAddr, config.ProvisionHostname = origAddr, origHostname
	})

	tests := []struct {
		name     string
		listen   string
		hostname string
		host     string
		want     string
	}{
		{name: "dhcp bare ip gets listen port", listen: "0.0.0.0:8081", host: "127.0.0.1", want: "http://127.0.0.1:8081"},
		{name: "host header keeps its port", listen: "0.0.0.0:8081", host: "127.0.0.1:8081", want: "http://127.0.0.1:8081"},
		{name: "host header port wins over listen port", listen: "0.0.0.0:8081", host: "127.0.0.1:9000", want: "http://127.0.0.1:9000"},
		{name: "host header ipv6 keeps its port", listen: "[::]:8081", host: "[::1]:8081", want: "http://[::1]:8081"},
		{name: "host header ipv6 without port gets listen port", listen: "[::]:8081", host: "[::1]", want: "http://[::1]:8081"},
		{name: "port 80 bare ip", listen: "0.0.0.0:80", host: "127.0.0.1", want: "http://127.0.0.1"},
		{name: "port 80 host header", listen: "0.0.0.0:80", host: "grendel.local", want: "http://grendel.local"},
		{name: "port 443 bare ip", listen: "0.0.0.0:443", host: "127.0.0.1", want: "http://127.0.0.1"},
		{name: "hostname replaces host header and gets listen port", listen: "0.0.0.0:8081", hostname: "grendel.local", host: "127.0.0.1:8081", want: "http://grendel.local:8081"},
		{name: "hostname with its own port keeps it", listen: "0.0.0.0:8081", hostname: "grendel.local:9000", host: "127.0.0.1:8081", want: "http://grendel.local:9000"},
		{name: "hostname on port 80", listen: "0.0.0.0:80", hostname: "grendel.local", host: "127.0.0.1", want: "http://grendel.local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.ProvisionAddr = netip.MustParseAddrPort(tt.listen)
			config.ProvisionHostname = tt.hostname

			e := NewEndpoints(tt.host, "token")
			assert.Equal(t, tt.want, e.BaseURL())
			assert.Equal(t, tt.want+"/repo", e.RepoURL())
			assert.Equal(t, tt.want+"/boot/token/ipxe", e.IpxeURL())
		})
	}
}
