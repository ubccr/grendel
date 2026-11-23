// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/config"
	"github.com/ubccr/grendel/internal/ping"
	"github.com/ubccr/grendel/pkg/model"
	"go4.org/netipx"
)

const (
	// smallest prefix length a range may have, so a search never walks more than 65536 addresses
	minRangeBits     = 16
	defaultNextCount = 3
	// maxNextCount bounds a pinged search, which pings pingSpare more than count for each list; an unpinged one, as a script adding a rack of nodes makes, may ask for up to maxNextCountUnpinged
	maxNextCount         = 8
	maxNextCountUnpinged = 64
	// pingSpare is how many addresses past count are pinged for each suggestion list, so a few that answer still leave count to offer
	pingSpare = 12
	pingTries = 2
	// a host on the same network answers in well under a millisecond, so this leaves room for a slow one while keeping a search to about half a second
	pingWait = 250 * time.Millisecond
)

// pingAddrs is ping.Ping, swapped out in tests
var pingAddrs = ping.Ping

type AddressUse struct {
	Addr netip.Addr `json:"ip" oai3:"typeStr"`
	// Bits is the prefix length the port was given with its address
	Bits int    `json:"-"`
	Node string `json:"node"`
	BMC  bool   `json:"bmc"`
}

// hostPorts is every port of host with an IPv4 address, its interfaces and its bonds
func hostPorts(host *model.Host) []*model.NetInterface {
	ports := slices.Clone(host.Interfaces)
	for _, b := range host.Bonds {
		ports = append(ports, &b.NetInterface)
	}
	return slices.DeleteFunc(ports, func(p *model.NetInterface) bool {
		return !p.IP.IsValid() || !p.IP.Addr().Is4()
	})
}

func addressUses(hosts model.HostList) []AddressUse {
	var uses []AddressUse
	for _, host := range hosts {
		for _, p := range hostPorts(host) {
			uses = append(uses, AddressUse{Addr: p.IP.Addr(), Bits: p.IP.Bits(), Node: host.Name, BMC: p.BMC})
		}
	}
	return uses
}

type NextAddress struct {
	IP       string `json:"ip" example:"10.17.25.7"`
	Answered bool   `json:"answered" description:"something answered a ping at the address, though no node holds it"`
}

type NextAddressResponse struct {
	Range        string        `json:"range" example:"10.17.25.0/24"`
	Subnet       string        `json:"subnet" description:"the subnet the range lies in: the one in dhcp.subnets holding it, or else taken from the prefix length of the nodes in and around it, or the range itself when there are none" example:"10.17.0.0/18"`
	PrefixLen    int           `json:"prefix_len" description:"the prefix length to give a port with one of the suggested addresses" example:"18"`
	Free         int           `json:"free" description:"how many addresses in the range no node holds, that aren't excluded or reserved"`
	Used         []AddressUse  `json:"used" description:"addresses in the range a node's port holds, lowest first"`
	Highest      string        `json:"highest,omitempty" description:"the highest address in the range a node holds or that is excluded, which after_highest starts from"`
	Reserved     []string      `json:"reserved" description:"addresses in the range never suggested: the subnet's network and broadcast addresses, the gateway DHCP hands its nodes, and in a subnet larger than a /24, ones ending in .0 or .255"`
	Lowest       []NextAddress `json:"lowest" description:"the lowest free addresses in the range"`
	AfterHighest []NextAddress `json:"after_highest" description:"the free addresses after the highest one in use"`
	Pinged       bool          `json:"pinged"`
	PingError    string        `json:"ping_error,omitempty"`
}

// parseRange reads the range to search: a CIDR, kept to its network, or an address without a prefix, or the first three octets of one, standing for its /24. It must be IPv4 and no larger than a /16
func parseRange(s string) (netip.Prefix, error) {
	s = strings.TrimSpace(s)
	var r netip.Prefix
	if strings.Contains(s, "/") {
		p, err := netip.ParsePrefix(s)
		if err != nil {
			return r, err
		}
		r = p.Masked()
	} else {
		if strings.Count(s, ".") == 2 {
			s += ".0"
		}
		a, err := netip.ParseAddr(s)
		if err != nil {
			return r, err
		}
		r = netip.PrefixFrom(a, 24).Masked()
	}
	if !r.Addr().Is4() {
		return r, errors.New("not an IPv4 range")
	}
	if r.Bits() < minRangeBits {
		return r, fmt.Errorf("larger than a /%d", minRangeBits)
	}
	return r, nil
}

// NodeNextAddress suggests free addresses in a range: the lowest ones no node holds, and the ones after the highest held, or only the list from names. An address that answers a ping is still listed, marked, so the caller can see what was skipped
func (h *Handler) NodeNextAddress(c fuego.ContextNoBody) (*NextAddressResponse, error) {
	cidr := c.QueryParam("cidr")
	r, err := parseRange(cidr)
	if err != nil {
		return nil, badRequest(fmt.Errorf("parsing range %q: %w", cidr, err), fmt.Sprintf("cidr must be an IPv4 range no larger than a /%d, like 10.17.25.0/24, an address standing for its /24, or three octets like 10.17.25", minRangeBits))
	}

	from := c.QueryParam("from")
	wantLowest, wantHigh := from == "" || from == "lowest", from == "" || from == "after_highest"
	if !wantLowest && !wantHigh {
		return nil, badRequest(fmt.Errorf("bad from %q", from), "from must be lowest or after_highest")
	}

	doPing, err := c.QueryParamBoolErr("ping")
	if err != nil {
		return nil, badRequest(err, "ping must be true or false")
	}
	most := maxNextCountUnpinged
	if doPing {
		most = maxNextCount
	}
	count, err := c.QueryParamIntErr("count")
	if err == nil && (count < 1 || count > most) {
		err = fmt.Errorf("not from 1 to %d", most)
	}
	if err != nil {
		return nil, badRequest(fmt.Errorf("bad count %q: %w", c.QueryParam("count"), err), fmt.Sprintf("count must be a number from 1 to %d, or to %d without pinging", maxNextCount, maxNextCountUnpinged))
	}

	// searching is rare enough that loading every node is simpler than a query for just their addresses
	hosts, err := h.DB.Hosts()
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to load nodes"}
	}

	p := planRange(r, addressUses(hosts), parseExclude(c.QueryParam("exclude")))
	res := p.response()

	n := count
	if doPing {
		n += pingSpare
	}
	var low, high []netip.Addr
	if wantLowest {
		low = p.lowest(n)
	}
	if wantHigh {
		high = p.afterHighest(n)
	}

	if !doPing {
		res.Lowest = unpinged(low)
		res.AfterHighest = unpinged(high)
		return res, nil
	}

	ctx, cancel := context.WithTimeout(c.Context(), pingTries*pingWait+time.Second)
	defer cancel()
	answered, err := pingAddrs(ctx, union(low, high), pingTries, pingWait)
	if err != nil {
		res.PingError = err.Error()
		res.Lowest = unpinged(low[:min(count, len(low))])
		res.AfterHighest = unpinged(high[:min(count, len(high))])
		return res, nil
	}

	res.Pinged = true
	res.Lowest = suggest(low, answered, count)
	res.AfterHighest = suggest(high, answered, count)
	return res, nil
}

type rangePlan struct {
	r    netip.Prefix
	bits int
	used map[netip.Addr]AddressUse
	// excluded is the map planRange was given, so it can hold addresses outside the range, and a caller sharing it between ranges sees what each adds
	excluded map[netip.Addr]bool
	gateways gateways
}

func planRange(r netip.Prefix, uses []AddressUse, exclude map[netip.Addr]bool) rangePlan {
	p := rangePlan{r: r, used: map[netip.Addr]AddressUse{}, excluded: exclude, gateways: configuredGateways()}
	if p.excluded == nil {
		p.excluded = map[netip.Addr]bool{}
	}
	for _, u := range uses {
		if r.Contains(u.Addr) {
			p.used[u.Addr] = u
		}
	}
	p.bits = subnetBits(r, uses)
	return p
}

// gateways is what decides the router DHCP hands a node, the same settings NetInterface.Gateway reads for a stored port
type gateways struct {
	// subnets are the dhcp.subnets entries, each a gateway address with its subnet's prefix length
	subnets []netip.Prefix
	// octet4 is dhcp.router_octet4, or -1 when it isn't set
	octet4 int
	static netip.Addr
}

func configuredGateways() gateways {
	g := gateways{octet4: -1, static: config.DefaultGateway}
	for _, s := range config.Subnets {
		g.subnets = append(g.subnets, s.Gateway)
	}
	if viper.IsSet("dhcp.router_octet4") {
		g.octet4 = viper.GetInt("dhcp.router_octet4")
	}
	return g
}

// has is true when a is a configured subnet's gateway, or the router a port with that address and prefix length would be handed. An address inside a configured subnet has no other router, as dhcp.router_octet4 and dhcp.gateway only apply outside them
func (g gateways) has(a netip.Addr, bits int) bool {
	configured := false
	for _, s := range g.subnets {
		if s.Addr() == a {
			return true
		}
		configured = configured || s.Contains(a)
	}
	if configured {
		return false
	}
	if g.octet4 >= 0 {
		last := netipx.PrefixLastIP(netip.PrefixFrom(a, bits)).As4()
		last[3] = uint8(g.octet4)
		return netip.AddrFrom4(last) == a
	}
	return g.static == a
}

// configuredBits is the prefix length of the smallest dhcp.subnets entry that holds the whole range
func configuredBits(r netip.Prefix) (int, bool) {
	best := -1
	for _, s := range config.Subnets {
		if s.Gateway.Bits() <= r.Bits() && s.Gateway.Contains(r.Addr()) {
			best = max(best, s.Gateway.Bits())
		}
	}
	return best, best >= 0
}

// subnetBits is the prefix length of the subnet a range lies in: that of a dhcp.subnets entry holding the range, else the most common among the nodes in the range if it holds the whole range. An empty range, often one rack's /24 of a larger subnet, takes the most common among nearby nodes whose subnet holds it. Failing those it is the most common among the range's nodes, or the range itself when it has none
func subnetBits(r netip.Prefix, uses []AddressUse) int {
	if b, ok := configuredBits(r); ok {
		return b
	}
	in := map[int]int{}
	around := map[int]int{}
	small := map[int]int{}
	for _, u := range uses {
		if u.Bits <= 0 || u.Bits > 32 || !u.Addr.Is4() {
			continue
		}
		holds := u.Bits <= r.Bits() && netip.PrefixFrom(u.Addr, u.Bits).Masked().Contains(r.Addr())
		switch {
		case holds && r.Contains(u.Addr):
			in[u.Bits]++
		case holds:
			around[u.Bits]++
		case r.Contains(u.Addr):
			small[u.Bits]++
		}
	}
	for _, counts := range []map[int]int{in, around, small} {
		if b, ok := mostCommon(counts); ok {
			return b
		}
	}
	return r.Bits()
}

// mostCommon picks the prefix length with the most nodes, the shorter one on a tie
func mostCommon(counts map[int]int) (int, bool) {
	best, most := 0, 0
	for b, n := range counts {
		if n > most || (n == most && b < best) {
			best, most = b, n
		}
	}
	return best, most > 0
}

// reserved is true for an address no port should be given: the subnet's gateway, its network and broadcast address, and in a subnet larger than a /24 any address ending in .0 or .255, which are valid there but look like network and broadcast addresses to some devices. A /31 or /32 has no network or broadcast address
func (p rangePlan) reserved(a netip.Addr) bool {
	if p.gateways.has(a, p.bits) {
		return true
	}
	if p.bits > 30 {
		return false
	}
	subnet := netip.PrefixFrom(a, p.bits).Masked()
	if a == subnet.Addr() || a == netipx.PrefixLastIP(subnet) {
		return true
	}
	last := a.As4()[3]
	return p.bits < 24 && (last == 0 || last == 255)
}

func (p rangePlan) free(a netip.Addr) bool {
	_, used := p.used[a]
	return !used && !p.excluded[a] && !p.reserved(a)
}

func (p rangePlan) lowest(n int) []netip.Addr {
	return p.freeFrom(p.r.Addr(), n)
}

// highest is the highest address in the range taken, by a node or by being excluded, or the zero Addr with none taken
func (p rangePlan) highest() netip.Addr {
	var top netip.Addr
	for a := range p.used {
		top = maxAddr(top, a)
	}
	for a := range p.excluded {
		if p.r.Contains(a) {
			top = maxAddr(top, a)
		}
	}
	return top
}

func maxAddr(a, b netip.Addr) netip.Addr {
	if !a.IsValid() || a.Less(b) {
		return b
	}
	return a
}

// afterHighest is the first n free addresses after the highest one taken, which with nothing taken is the lowest
func (p rangePlan) afterHighest(n int) []netip.Addr {
	top := p.highest()
	if !top.IsValid() {
		return p.lowest(n)
	}
	return p.freeFrom(top.Next(), n)
}

func (p rangePlan) freeFrom(a netip.Addr, n int) []netip.Addr {
	var out []netip.Addr
	for ; len(out) < n && a.IsValid() && p.r.Contains(a); a = a.Next() {
		if p.free(a) {
			out = append(out, a)
		}
	}
	return out
}

func (p rangePlan) response() *NextAddressResponse {
	res := &NextAddressResponse{
		Range:     p.r.String(),
		Subnet:    netip.PrefixFrom(p.r.Addr(), p.bits).Masked().String(),
		PrefixLen: p.bits,
		Used:      []AddressUse{},
		Reserved:  []string{},
	}
	if p.bits > p.r.Bits() {
		// the range holds several subnets, so it is named as the subnet
		res.Subnet = p.r.String()
	}
	if top := p.highest(); top.IsValid() {
		res.Highest = top.String()
	}

	for _, a := range slices.SortedFunc(maps.Keys(p.used), netip.Addr.Compare) {
		res.Used = append(res.Used, p.used[a])
	}

	for a := p.r.Addr(); a.IsValid() && p.r.Contains(a); a = a.Next() {
		switch {
		case p.reserved(a):
			res.Reserved = append(res.Reserved, a.String())
		case p.free(a):
			res.Free++
		}
	}
	return res
}

// suggest takes addresses in order until n have gone unanswered, keeping the ones that answered on the way
func suggest(addrs []netip.Addr, answered map[netip.Addr]bool, n int) []NextAddress {
	out := []NextAddress{}
	quiet := 0
	for _, a := range addrs {
		if quiet == n {
			break
		}
		out = append(out, NextAddress{IP: a.String(), Answered: answered[a]})
		if !answered[a] {
			quiet++
		}
	}
	return out
}

func unpinged(addrs []netip.Addr) []NextAddress {
	out := make([]NextAddress, 0, len(addrs))
	for _, a := range addrs {
		out = append(out, NextAddress{IP: a.String()})
	}
	return out
}

// union is a and b without repeats, in address order
func union(a, b []netip.Addr) []netip.Addr {
	out := slices.Concat(a, b)
	slices.SortFunc(out, netip.Addr.Compare)
	return slices.Compact(out)
}

// parseExclude reads a comma separated list of addresses, each with or without a prefix length, as the IP field of a port holds them. What doesn't parse is ignored
func parseExclude(s string) map[netip.Addr]bool {
	out := map[netip.Addr]bool{}
	for _, f := range strings.Split(s, ",") {
		addr, _, _ := strings.Cut(strings.TrimSpace(f), "/")
		if a, err := netip.ParseAddr(addr); err == nil {
			out[a] = true
		}
	}
	return out
}
