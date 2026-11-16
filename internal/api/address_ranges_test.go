// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net/netip"
	"testing"

	"github.com/go-fuego/fuego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
)

// storeRacked stores a node with a boot port and a BMC port, tagged into rack
func storeRacked(t *testing.T, h *Handler, name, rack, boot, bmc string) {
	t.Helper()
	host := tests.HostFactory.MustCreate().(*model.Host)
	host.Name = name
	host.Tags = []string{"rack=" + rack}
	host.Interfaces[0].IP = netip.MustParsePrefix(boot)
	host.Interfaces[1].IP = netip.MustParsePrefix(bmc)
	host.Bonds[0].IP = netip.Prefix{}
	require.NoError(t, h.DB.StoreHost(host))
}

func addressRangesOf(t *testing.T, h *Handler, params map[string]string) []AddressRange {
	t.Helper()
	c := fuego.NewMockContextNoBody()
	for _, name := range []string{"nodeset", "tags", "bmc"} {
		c.OpenAPIParams[name] = fuego.OpenAPIParam{Name: name}
	}
	for k, v := range params {
		c.SetQueryParam(k, v)
	}
	res, err := h.NodeAddressRanges(c)
	require.NoError(t, err)
	return res
}

func rangeNames(list []AddressRange) []string {
	out := []string{}
	for _, r := range list {
		out = append(out, r.Range)
	}
	return out
}

func TestNodeAddressRanges(t *testing.T) {
	h := newTestHandler(t)
	storeRacked(t, h, "cpn-u25-01", "u25", "10.17.25.1/18", "10.18.25.1/18")
	storeRacked(t, h, "cpn-u25-02", "u25", "10.17.25.2/18", "10.18.25.2/18")
	storeRacked(t, h, "cpn-u26-01", "u26", "10.17.26.1/18", "10.18.0.1/32")

	t.Run("every node, in address order, with what each range holds", func(t *testing.T) {
		res := addressRangesOf(t, h, nil)
		assert.Equal(t, []AddressRange{
			{Range: "10.17.25.0/24", PrefixLen: 18, Ports: 2, BMCPorts: 0, Nodes: 2},
			{Range: "10.17.26.0/24", PrefixLen: 18, Ports: 1, BMCPorts: 0, Nodes: 1},
			{Range: "10.18.0.0/24", PrefixLen: 32, Ports: 1, BMCPorts: 1, Nodes: 1},
			{Range: "10.18.25.0/24", PrefixLen: 18, Ports: 2, BMCPorts: 2, Nodes: 2},
		}, res)
	})

	t.Run("by tag", func(t *testing.T) {
		res := addressRangesOf(t, h, map[string]string{"tags": "rack=u25"})
		assert.Equal(t, []string{"10.17.25.0/24", "10.18.25.0/24"}, rangeNames(res))
	})

	t.Run("by nodeset and by part of a name", func(t *testing.T) {
		assert.Equal(t, []string{"10.17.26.0/24", "10.18.0.0/24"}, rangeNames(addressRangesOf(t, h, map[string]string{"nodeset": "cpn-u26-[01-02]"})))
		assert.Equal(t, []string{"10.17.26.0/24", "10.18.0.0/24"}, rangeNames(addressRangesOf(t, h, map[string]string{"nodeset": "U26"})))
		assert.Equal(t, []string{"10.17.25.0/24", "10.18.25.0/24"}, rangeNames(addressRangesOf(t, h, map[string]string{"nodeset": "cpn-u25-[0"})))
	})

	t.Run("by tags, any node with one of them as elsewhere in the API", func(t *testing.T) {
		res := addressRangesOf(t, h, map[string]string{"tags": "rack=u26,rack=u99"})
		assert.Equal(t, []string{"10.17.26.0/24", "10.18.0.0/24"}, rangeNames(res))
		assert.Empty(t, addressRangesOf(t, h, map[string]string{"tags": "rack=u99"}))
	})

	t.Run("by both, which must both match", func(t *testing.T) {
		assert.Empty(t, addressRangesOf(t, h, map[string]string{"nodeset": "cpn-u26", "tags": "rack=u25"}))
	})

	t.Run("the kind of port asked for first, then the busiest", func(t *testing.T) {
		res := addressRangesOf(t, h, map[string]string{"bmc": "true"})
		assert.Equal(t, []string{"10.18.25.0/24", "10.18.0.0/24", "10.17.25.0/24", "10.17.26.0/24"}, rangeNames(res))
	})
}

func TestNodeFilter(t *testing.T) {
	names := []string{"cpn-u25-01", "cpn-u25-02", "cpn-u25-10", "cpn-u26-01", "srv-u25-01"}
	pick := func(text string) []string {
		named := nodeFilter(text)
		out := []string{}
		for _, n := range names {
			if named(n) {
				out = append(out, n)
			}
		}
		return out
	}

	assert.Equal(t, []string{"cpn-u25-01", "cpn-u25-02"}, pick("cpn-u25-[01-02]"))
	assert.Equal(t, []string{"cpn-u25-10", "srv-u25-01"}, pick("cpn-u25-10,srv-u25-01"))
	assert.Equal(t, []string{"cpn-u25-01", "srv-u25-01"}, pick("U25-01"))
	assert.Equal(t, []string{"cpn-u25-01", "cpn-u25-02", "cpn-u25-10"}, pick("cpn-u25-[0"))
	assert.Equal(t, names, pick(" "))
}
