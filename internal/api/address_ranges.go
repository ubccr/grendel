// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"cmp"
	"errors"
	"maps"
	"net/netip"
	"slices"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

type AddressRange struct {
	Range     string `json:"range" example:"10.17.25.0/24"`
	PrefixLen int    `json:"prefix_len" description:"the prefix length most of the range's ports carry" example:"18"`
	Ports     int    `json:"ports" description:"how many of the nodes' ports have an address in the range"`
	BMCPorts  int    `json:"bmc_ports" description:"how many of those ports are BMCs"`
	Nodes     int    `json:"nodes" description:"how many nodes those ports belong to"`
}

// NodeAddressRanges lists the ranges the ports of some nodes sit in, each once however many ports share it, so a range for a new node can be found from where similar nodes are. Nodes are picked by name, by tags, or both, and every node is taken with neither
func (h *Handler) NodeAddressRanges(c fuego.ContextNoBody) ([]AddressRange, error) {
	hosts, err := h.DB.Hosts()
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to load nodes"}
	}

	named := nodeFilter(c.QueryParam("nodeset"))
	tagged := func(string) bool { return true }
	if t := strings.TrimSpace(c.QueryParam("tags")); t != "" {
		// tags pick nodes as they do everywhere else in the API, any node with one of them
		ns, err := h.DB.FindTags(strings.Split(t, ","))
		if errors.Is(err, store.ErrNotFound) {
			return []AddressRange{}, nil
		}
		if err != nil {
			return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to find nodes by tags"}
		}
		names := ns.Iterator().StringSlice()
		tagged = func(name string) bool { return slices.Contains(names, name) }
	}
	picked := slices.DeleteFunc(hosts, func(host *model.Host) bool {
		return !named(host.Name) || !tagged(host.Name)
	})

	// a missing bmc leaves the ranges in address order rather than putting one kind of port first
	bmc, err := c.QueryParamBoolErr("bmc")
	return addressRanges(picked, bmc, err == nil), nil
}

// rangeOf is the range an address is searched in: its own subnet when that is from a /24 to a /30, and otherwise the /24 it sits in, as a rack's share of a larger subnet usually is. A /31 or /32 is a link or a single host rather than a subnet to fill, so it too gives its /24
func rangeOf(p netip.Prefix) netip.Prefix {
	bits := p.Bits()
	if bits < 24 || bits > 30 {
		bits = 24
	}
	return netip.PrefixFrom(p.Addr(), bits).Masked()
}

// addressRanges groups the ports of hosts by rangeOf their address. With sortByKind, the ranges holding the kind of port asked for, BMC or not, come first, then the busiest; otherwise they are in address order
func addressRanges(hosts model.HostList, bmc, sortByKind bool) []AddressRange {
	type group struct {
		AddressRange
		nodes    map[string]bool
		prefixes map[int]int
	}
	groups := map[netip.Prefix]*group{}
	for _, host := range hosts {
		for _, port := range hostPorts(host) {
			r := rangeOf(port.IP)
			g := groups[r]
			if g == nil {
				g = &group{AddressRange: AddressRange{Range: r.String()}, nodes: map[string]bool{}, prefixes: map[int]int{}}
				groups[r] = g
			}
			g.Ports++
			if port.BMC {
				g.BMCPorts++
			}
			g.nodes[host.Name] = true
			g.prefixes[port.IP.Bits()]++
		}
	}

	out := make([]AddressRange, 0, len(groups))
	for _, r := range slices.SortedFunc(maps.Keys(groups), netip.Prefix.Compare) {
		g := groups[r]
		g.Nodes = len(g.nodes)
		g.PrefixLen, _ = mostCommon(g.prefixes)
		out = append(out, g.AddressRange)
	}
	if sortByKind {
		kind := func(r AddressRange) bool {
			if bmc {
				return r.BMCPorts > 0
			}
			return r.Ports > r.BMCPorts
		}
		// stable, so ranges alike in both stay in address order
		slices.SortStableFunc(out, func(a, b AddressRange) int {
			if ka, kb := kind(a), kind(b); ka != kb {
				if ka {
					return -1
				}
				return 1
			}
			return cmp.Compare(b.Ports, a.Ports)
		})
	}
	return out
}

// nodeFilter reads text picking nodes by name. A nodeset with brackets or commas, like cpn-u25-[01-40], picks the nodes it names; plain text, or a nodeset not finished yet, picks the names holding the text before its first bracket, so cpn-u25 or cpn-u25-[0 narrow the nodes as they are typed. Empty text picks every node
func nodeFilter(text string) func(name string) bool {
	t := strings.ToLower(strings.TrimSpace(text))
	if t == "" {
		return func(string) bool { return true }
	}
	if strings.ContainsAny(t, "[,") {
		if ns, err := nodeset.NewNodeSet(t); err == nil && ns.Len() > 0 {
			names := map[string]bool{}
			for _, n := range ns.Iterator().StringSlice() {
				names[n] = true
			}
			return func(name string) bool { return names[strings.ToLower(name)] }
		}
	}
	part, _, _ := strings.Cut(t, "[")
	part, _, _ = strings.Cut(part, ",")
	part = strings.TrimSpace(part)
	return func(name string) bool { return strings.Contains(strings.ToLower(name), part) }
}
