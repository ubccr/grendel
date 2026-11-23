// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/config"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
)

func storeAddressed(t *testing.T, h *Handler, ips ...string) {
	t.Helper()
	for _, ip := range ips {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = "n-" + ip
		host.Interfaces[0].IP = netip.MustParsePrefix(ip)
		host.Interfaces[1].IP = netip.Prefix{}
		host.Bonds[0].IP = netip.Prefix{}
		require.NoError(t, h.DB.StoreHost(host))
	}
}

func fakePing(t *testing.T, answer ...string) *[]netip.Addr {
	t.Helper()
	var asked []netip.Addr
	old := pingAddrs
	pingAddrs = func(_ context.Context, addrs []netip.Addr, _ int, _ time.Duration) (map[netip.Addr]bool, error) {
		asked = append(asked, addrs...)
		out := map[netip.Addr]bool{}
		for _, a := range answer {
			out[netip.MustParseAddr(a)] = true
		}
		return out, nil
	}
	t.Cleanup(func() { pingAddrs = old })
	return &asked
}

func nextAddress(t *testing.T, h *Handler, params map[string]string) (*NextAddressResponse, error) {
	t.Helper()
	c := fuego.NewMockContextNoBody()
	// the route declares these defaults, which a mock context doesn't know
	c.OpenAPIParams["count"] = fuego.OpenAPIParam{Name: "count", Default: defaultNextCount}
	c.OpenAPIParams["ping"] = fuego.OpenAPIParam{Name: "ping", Default: true}
	c.OpenAPIParams["exclude"] = fuego.OpenAPIParam{Name: "exclude"}
	c.OpenAPIParams["from"] = fuego.OpenAPIParam{Name: "from"}
	for k, v := range params {
		c.SetQueryParam(k, v)
	}
	return h.NodeNextAddress(c)
}

func ips(list []NextAddress) []string {
	out := []string{}
	for _, a := range list {
		out = append(out, a.IP)
	}
	return out
}

func TestNextAddressRackOfLargerSubnet(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)
	// a rack's /24 of a /18: .0 and .255 are valid hosts in a /18 but are never offered
	storeAddressed(t, h, "10.17.25.1/18", "10.17.25.2/18", "10.17.25.5/18", "10.17.24.9/18")

	res, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24"})
	require.NoError(t, err)

	assert.Equal(t, "10.17.0.0/18", res.Subnet)
	assert.Equal(t, 18, res.PrefixLen)
	assert.Equal(t, 251, res.Free, "256 less 3 nodes, and .0 and .255")
	assert.Equal(t, "10.17.25.5", res.Highest)
	assert.Equal(t, []string{"10.17.25.0", "10.17.25.255"}, res.Reserved)
	assert.Equal(t, []string{"10.17.25.3", "10.17.25.4", "10.17.25.6"}, ips(res.Lowest))
	assert.Equal(t, []string{"10.17.25.6", "10.17.25.7", "10.17.25.8"}, ips(res.AfterHighest))
	assert.Len(t, res.Used, 3, "a node outside the range isn't listed")
	assert.True(t, res.Pinged)
}

func TestNextAddressEmptyRackTakesSubnetFromNeighbours(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)
	storeAddressed(t, h, "10.17.24.9/18")

	res, err := nextAddress(t, h, map[string]string{"cidr": "10.17.30.0/24", "count": "1"})
	require.NoError(t, err)

	assert.Equal(t, 18, res.PrefixLen)
	assert.Equal(t, []string{"10.17.30.1"}, ips(res.Lowest))
	assert.Equal(t, []string{"10.17.30.1"}, ips(res.AfterHighest), "with nothing used, after the highest starts at the lowest")
}

func TestNextAddressNoNodesUsesRange(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)

	res, err := nextAddress(t, h, map[string]string{"cidr": "192.168.7.64/26", "count": "1"})
	require.NoError(t, err)

	assert.Equal(t, "192.168.7.64/26", res.Subnet)
	assert.Equal(t, []string{"192.168.7.64", "192.168.7.127"}, res.Reserved)
	assert.Equal(t, []string{"192.168.7.65"}, ips(res.Lowest))
}

func TestNextAddressKeepsAnsweredAddresses(t *testing.T) {
	h := newTestHandler(t)
	asked := fakePing(t, "10.17.25.3")
	storeAddressed(t, h, "10.17.25.1/24", "10.17.25.2/24")

	res, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "count": "2"})
	require.NoError(t, err)

	assert.Equal(t, []NextAddress{
		{IP: "10.17.25.3", Answered: true},
		{IP: "10.17.25.4"},
		{IP: "10.17.25.5"},
	}, res.Lowest)
	assert.Contains(t, *asked, netip.MustParseAddr("10.17.25.3"))
	assert.NotContains(t, *asked, netip.MustParseAddr("10.17.25.1"), "a node's address isn't pinged")
}

func TestNextAddressExclude(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)
	storeAddressed(t, h, "10.17.25.1/24")

	res, err := nextAddress(t, h, map[string]string{
		"cidr":    "10.17.25.0/24",
		"count":   "1",
		"exclude": "10.17.25.2/24, 10.17.25.10,junk,10.99.0.1",
	})
	require.NoError(t, err)

	assert.Equal(t, "10.17.25.10", res.Highest, "an excluded address counts as taken")
	assert.Equal(t, 251, res.Free, "254 less the node and the 2 excluded addresses")
	assert.Equal(t, []string{"10.17.25.3"}, ips(res.Lowest))
	assert.Equal(t, []string{"10.17.25.11"}, ips(res.AfterHighest))
}

func TestNextAddressFullRange(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)
	storeAddressed(t, h, "10.0.0.1/30", "10.0.0.2/30")

	res, err := nextAddress(t, h, map[string]string{"cidr": "10.0.0.0/30"})
	require.NoError(t, err)

	assert.Empty(t, res.Lowest)
	assert.Empty(t, res.AfterHighest)
}

func TestNextAddressWithoutPing(t *testing.T) {
	h := newTestHandler(t)
	asked := fakePing(t)

	res, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "ping": "false", "count": "2"})
	require.NoError(t, err)

	assert.False(t, res.Pinged)
	assert.Equal(t, []string{"10.17.25.1", "10.17.25.2"}, ips(res.Lowest))
	assert.Empty(t, *asked)
}

func TestNextAddressPingFails(t *testing.T) {
	h := newTestHandler(t)
	old := pingAddrs
	pingAddrs = func(context.Context, []netip.Addr, int, time.Duration) (map[netip.Addr]bool, error) {
		return nil, errors.New("operation not permitted")
	}
	t.Cleanup(func() { pingAddrs = old })

	res, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "count": "2"})
	require.NoError(t, err, "a failed ping still suggests addresses")

	assert.False(t, res.Pinged)
	assert.Equal(t, "operation not permitted", res.PingError)
	assert.Equal(t, []string{"10.17.25.1", "10.17.25.2"}, ips(res.Lowest))
}

func TestNextAddressRefusesBadRanges(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)

	for _, cidr := range []string{"", "10.17", "fd00::/64", "fd00::1", "10.0.0.0/8", "10.17.25.0/33"} {
		_, err := nextAddress(t, h, map[string]string{"cidr": cidr})
		var httpErr fuego.HTTPError
		if assert.ErrorAs(t, err, &httpErr, cidr) {
			assert.Equal(t, 400, httpErr.StatusCode(), cidr)
		}
	}

	for _, count := range []string{"0", "9", "x"} {
		_, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "count": count})
		assert.Error(t, err, count)
	}

	_, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "ping": "maybe"})
	assert.Error(t, err)
}

func TestNextAddressReadsPingAsBool(t *testing.T) {
	h := newTestHandler(t)
	asked := fakePing(t)

	for _, off := range []string{"false", "0", "F"} {
		res, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "ping": off})
		require.NoError(t, err)
		assert.False(t, res.Pinged, off)
	}
	assert.Empty(t, *asked)
}

func TestNextAddressPrefersNodesInTheRange(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)
	// the neighbours outside the range use a different prefix length than the nodes in it, and outnumber them
	storeAddressed(t, h, "10.17.25.1/18", "10.17.26.1/16", "10.17.27.1/16", "10.17.28.1/16")

	res, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "count": "1"})
	require.NoError(t, err)
	assert.Equal(t, 18, res.PrefixLen)
}

// withSubnets sets dhcp.subnets, each given as a gateway with its subnet's prefix length, until the test ends
func withSubnets(t *testing.T, gateways ...string) {
	t.Helper()
	old := config.Subnets
	t.Cleanup(func() { config.Subnets = old })
	config.Subnets = nil
	for _, g := range gateways {
		config.Subnets = append(config.Subnets, config.Subnet{Gateway: netip.MustParsePrefix(g)})
	}
}

func TestNextAddressConfiguredSubnetDecides(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)
	withSubnets(t, "10.17.0.1/16", "10.17.25.254/18")
	// the nodes carry a prefix length the configured subnet disagrees with
	storeAddressed(t, h, "10.17.25.1/24", "10.17.25.2/24")

	res, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "count": "1"})
	require.NoError(t, err)

	assert.Equal(t, 18, res.PrefixLen, "the smallest configured subnet holding the range")
	assert.Equal(t, "10.17.0.0/18", res.Subnet)
	assert.Equal(t, []string{"10.17.25.0", "10.17.25.254", "10.17.25.255"}, res.Reserved)
	assert.Equal(t, 251, res.Free, "256 less 2 nodes, the gateway, and .0 and .255")
}

func TestNextAddressNeverSuggestsTheGateway(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)

	t.Run("of a configured subnet", func(t *testing.T) {
		withSubnets(t, "10.9.0.1/24")
		res, err := nextAddress(t, h, map[string]string{"cidr": "10.9.0.0/24", "count": "1", "ping": "false"})
		require.NoError(t, err)
		assert.Equal(t, []string{"10.9.0.0", "10.9.0.1", "10.9.0.255"}, res.Reserved)
		assert.Equal(t, []string{"10.9.0.2"}, ips(res.Lowest))
	})

	t.Run("from dhcp.router_octet4", func(t *testing.T) {
		viper.Set("dhcp.router_octet4", 1)
		t.Cleanup(func() { viper.Set("dhcp.router_octet4", nil) })
		res, err := nextAddress(t, h, map[string]string{"cidr": "10.9.0.0/24", "count": "1", "ping": "false"})
		require.NoError(t, err)
		assert.Equal(t, []string{"10.9.0.0", "10.9.0.1", "10.9.0.255"}, res.Reserved)
		assert.Equal(t, []string{"10.9.0.2"}, ips(res.Lowest))
	})

	t.Run("a configured subnet's gateway wins over dhcp.router_octet4, as it does for a node", func(t *testing.T) {
		withSubnets(t, "10.9.0.254/24")
		viper.Set("dhcp.router_octet4", 1)
		t.Cleanup(func() { viper.Set("dhcp.router_octet4", nil) })
		res, err := nextAddress(t, h, map[string]string{"cidr": "10.9.0.0/24", "count": "1", "ping": "false"})
		require.NoError(t, err)
		assert.Equal(t, []string{"10.9.0.0", "10.9.0.254", "10.9.0.255"}, res.Reserved)
		assert.Equal(t, []string{"10.9.0.1"}, ips(res.Lowest))
	})
}

func TestNextAddressBadCountSaysWhy(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)

	_, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "count": "9"})
	var httpErr fuego.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, `bad count "9": not from 1 to 8`, httpErr.Err.Error())
}

func TestNextAddressTopOfRangeUsed(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)
	storeAddressed(t, h, "10.17.25.254/24")

	res, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "count": "1"})
	require.NoError(t, err)
	assert.Empty(t, res.AfterHighest, "nothing lies after the highest address in use")
	assert.Equal(t, []string{"10.17.25.1"}, ips(res.Lowest))
	assert.Equal(t, 253, res.Free, "which the count of free addresses still shows")
}

func TestNextAddressRangeOfSmallSubnets(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)
	storeAddressed(t, h, "10.17.25.1/26", "10.17.25.65/26")

	res, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "count": "1"})
	require.NoError(t, err)

	assert.Equal(t, 26, res.PrefixLen)
	assert.Equal(t, "10.17.25.0/24", res.Subnet)
	assert.Equal(t, []string{
		"10.17.25.0", "10.17.25.63", "10.17.25.64", "10.17.25.127",
		"10.17.25.128", "10.17.25.191", "10.17.25.192", "10.17.25.255",
	}, res.Reserved)
	assert.Equal(t, []string{"10.17.25.2"}, ips(res.Lowest))
	assert.Equal(t, []string{"10.17.25.66"}, ips(res.AfterHighest))
}

func TestNextAddressCountLimitDependsOnPinging(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)

	res, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "ping": "false", "count": "64"})
	require.NoError(t, err)
	assert.Len(t, res.Lowest, 64)

	_, err = nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "count": "9"})
	assert.Error(t, err, "a pinged search stays within 8")
	_, err = nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "ping": "false", "count": "65"})
	assert.Error(t, err)
}

func TestNextAddressShorthandRanges(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)

	for _, cidr := range []string{"10.17.25.0/24", "10.17.25.77/24", "10.17.25.7", "10.17.25", " 10.17.25 "} {
		res, err := nextAddress(t, h, map[string]string{"cidr": cidr, "count": "1"})
		require.NoError(t, err, cidr)
		assert.Equal(t, "10.17.25.0/24", res.Range, cidr)
	}
}

func TestNextAddressFrom(t *testing.T) {
	h := newTestHandler(t)
	asked := fakePing(t)
	storeAddressed(t, h, "10.17.25.1/24", "10.17.25.5/24")

	res, err := nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "count": "1", "from": "after_highest"})
	require.NoError(t, err)
	assert.Empty(t, res.Lowest)
	assert.Equal(t, []string{"10.17.25.6"}, ips(res.AfterHighest))
	assert.NotContains(t, *asked, netip.MustParseAddr("10.17.25.2"), "the list not asked for isn't pinged")

	res, err = nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "count": "1", "from": "lowest", "ping": "false"})
	require.NoError(t, err)
	assert.Equal(t, []string{"10.17.25.2"}, ips(res.Lowest))
	assert.Empty(t, res.AfterHighest)

	_, err = nextAddress(t, h, map[string]string{"cidr": "10.17.25.0/24", "from": "middle"})
	assert.Error(t, err)
}
