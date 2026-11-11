// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/pkg/model"
)

func host(name string, nics ...*model.NetInterface) *model.Host {
	return &model.Host{Name: name, Interfaces: nics}
}

func bmcNic(cidr string) *model.NetInterface {
	nic := &model.NetInterface{BMC: true}
	if cidr != "" {
		nic.IP = netip.MustParsePrefix(cidr)
	}
	return nic
}

func TestTargetsFromHosts(t *testing.T) {
	hosts := model.HostList{
		host("cpn-a01-01", bmcNic("10.1.0.1/16")),
		host("cpn-a01-02", &model.NetInterface{BMC: false, IP: netip.MustParsePrefix("10.2.0.2/16")}),
		host("pdu-f01-r", bmcNic("10.1.0.3/16")),
		host("swe-f01-r", &model.NetInterface{BMC: false, IP: netip.MustParsePrefix("10.2.0.4/16")}),
		host("cpn-a01-04", bmcNic("")),
	}

	targets := TargetsFromHosts(hosts)

	// the bmc flag decides and the name does not, in both directions
	require.Len(t, targets, 3)
	assert.Equal(t, Target{Name: "cpn-a01-01", Addr: "10.1.0.1"}, targets[0])
	assert.Equal(t, Target{Name: "pdu-f01-r", Addr: "10.1.0.3"}, targets[1])
	assert.Equal(t, Target{Name: "cpn-a01-04", Addr: ""}, targets[2], "connect reports the missing address, so the node still gets a row")
}

func TestEndpointScheme(t *testing.T) {
	assert.Equal(t, "https://10.1.0.1", endpoint("https", "10.1.0.1"))
	assert.Equal(t, "https://10.1.0.1:8443", endpoint("https", "10.1.0.1:8443"))
	assert.Equal(t, "http://127.0.0.1:1234", endpoint("https", "http://127.0.0.1:1234"), "tests point at an httptest server")
	assert.Equal(t, "http://10.1.0.1:8000", endpoint("http", "10.1.0.1:8000"), "bmc.scheme reaches a mockup server, which speaks plain http")
	assert.Equal(t, "https://10.1.0.1", endpoint("", "10.1.0.1"), "an unset scheme is https, not the endpoint ://10.1.0.1")
}
