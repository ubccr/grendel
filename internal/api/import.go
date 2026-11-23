// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/pkg/model"
)

const (
	ImportCreate    = "create"
	ImportAdd       = "add"
	ImportReplace   = "replace"
	ImportUnchanged = "unchanged"
	ImportConflict  = "conflict"

	maxImportRows = 5000
	maxImportPing = 2048
)

type NodeImportRow struct {
	Node   string   `json:"node" example:"cpn-a01-07"`
	MAC    string   `json:"mac" example:"d0:8e:79:00:00:07"`
	IP     string   `json:"ip,omitempty" description:"An address, with or without a prefix length. Left empty, one is taken from the range for its kind of port" example:"10.1.1.7/22"`
	FQDN   string   `json:"fqdn,omitempty" example:"bmc-cpn-a01-07.cluster"`
	Ifname string   `json:"ifname,omitempty"`
	BMC    bool     `json:"bmc" description:"A BMC row sets the node's one BMC port, replacing its MAC if it differs. Any other row updates the node's port with that MAC, or adds one after its others, so its boot port, the first that is not a BMC, stays first"`
	VLAN   string   `json:"vlan,omitempty"`
	MTU    uint16   `json:"mtu,omitempty"`
	Tags   []string `json:"tags,omitempty" oai3:"nullable"`
}

type NodeImportRanges struct {
	BMC  string `json:"bmc,omitempty" example:"10.1.0.0/22"`
	Boot string `json:"boot,omitempty" example:"10.2.0.0/22"`
}

type NodeImportRequest struct {
	DryRun bool             `json:"dry_run"`
	Ping   bool             `json:"ping" description:"Ping the addresses taken from a range and skip any that answer. Only done on a dry run"`
	From   string           `json:"from,omitempty" description:"Where in a range to start: lowest, the default, or after_highest" example:"lowest"`
	Ranges NodeImportRanges `json:"ranges"`
	Rows   []NodeImportRow  `json:"rows"`
}

type NodeImportResult struct {
	NodeImportRow
	Action   string `json:"action" description:"create, add, replace, unchanged or conflict"`
	Reason   string `json:"reason,omitempty"`
	Replaces string `json:"replaces,omitempty" description:"The MAC of the BMC port a replace swaps out"`
	Assigned bool   `json:"assigned" description:"The address was taken from a range rather than given"`
	Answered bool   `json:"answered,omitempty" description:"An address that was given answered a ping"`
}

type NodeImportSummary struct {
	Create    int `json:"create"`
	Add       int `json:"add"`
	Replace   int `json:"replace"`
	Unchanged int `json:"unchanged"`
	Conflict  int `json:"conflict"`
}

type NodeImportResponse struct {
	Rows      []NodeImportResult `json:"rows"`
	Summary   NodeImportSummary  `json:"summary"`
	Committed bool               `json:"committed"`
	Pinged    bool               `json:"pinged"`
	PingError string             `json:"ping_error,omitempty"`
}

// NodeImport adds a port to each node in the request, creating nodes that do not exist. A dry run reports what would change and fills in addresses from the ranges; a commit stores every row in one transaction, or nothing if any row conflicts
func (h *Handler) NodeImport(c fuego.ContextWithBody[NodeImportRequest]) (*NodeImportResponse, error) {
	req, err := c.Body()
	if err != nil {
		return nil, badRequest(err, "failed to parse the import request")
	}
	if len(req.Rows) == 0 {
		return nil, badRequest(errors.New("no rows"), "the import has no rows")
	}
	if len(req.Rows) > maxImportRows {
		return nil, badRequest(fmt.Errorf("%d rows", len(req.Rows)), fmt.Sprintf("an import may have at most %d rows", maxImportRows))
	}
	opts, err := importOptions(req)
	if err != nil {
		return nil, err
	}

	hosts, err := h.DB.Hosts()
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to load nodes"}
	}

	res := &NodeImportResponse{}
	var plan importPlan
	if req.DryRun && req.Ping {
		plan, res.Pinged, res.PingError = planWithPing(c.Context(), hosts, req.Rows, opts)
	} else {
		plan = planImport(hosts, req.Rows, opts)
	}
	res.Rows, res.Summary = plan.results(), plan.summary()

	if req.DryRun {
		return res, nil
	}
	if res.Summary.Conflict > 0 {
		return nil, fuego.HTTPError{
			Status: http.StatusConflict,
			Err:    fmt.Errorf("%d conflicting rows", res.Summary.Conflict),
			Title:  "Conflict",
			Detail: fmt.Sprintf("%d rows conflict with existing nodes or each other. Run a dry run to see them", res.Summary.Conflict),
		}
	}

	changed := plan.hosts()
	if len(changed) == 0 {
		res.Committed = true
		return res, nil
	}
	if err := h.DB.StoreHosts(changed); err != nil {
		return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: fmt.Sprintf("failed to store nodes: %s", err)}
	}
	if ns, err := changed.ToNodeSet(); err == nil {
		h.audit(c.Context(), model.SeveritySuccess, "node.import", fmt.Sprintf("Imported %d ports", len(plan.rows)-res.Summary.Unchanged), ns.String())
	}
	res.Committed = true
	return res, nil
}

type importOpts struct {
	bmc, boot    netip.Prefix
	afterHighest bool
}

func importOptions(req NodeImportRequest) (importOpts, error) {
	var o importOpts
	switch req.From {
	case "", "lowest":
	case "after_highest":
		o.afterHighest = true
	default:
		return o, badRequest(fmt.Errorf("bad from %q", req.From), "from must be lowest or after_highest")
	}
	for _, r := range []struct {
		s    string
		dst  *netip.Prefix
		name string
	}{{req.Ranges.BMC, &o.bmc, "bmc"}, {req.Ranges.Boot, &o.boot, "boot"}} {
		if r.s == "" {
			continue
		}
		p, err := parseRange(r.s)
		if err != nil {
			return o, badRequest(fmt.Errorf("parsing %s range %q: %w", r.name, r.s, err), fmt.Sprintf("the %s range must be an IPv4 range no larger than a /%d", r.name, minRangeBits))
		}
		*r.dst = p
	}
	return o, nil
}

type plannedRow struct {
	NodeImportResult
	mac net.HardwareAddr
	// ip is the address the row's port ends up with: the one given, the one the port has, or one taken from a range. It stays zero for a row that gets none, whatever text IP holds
	ip netip.Prefix
	// host is the stored node the row changes, nil for a node the import creates
	host *model.Host
	// port is the port of host the row changes, nil when the row adds one
	port *model.NetInterface
}

func (r *plannedRow) conflict(reason string) {
	if r.Action != ImportConflict {
		r.Action = ImportConflict
		r.Reason = reason
	}
}

func (r *plannedRow) setIP(p netip.Prefix) {
	r.ip = p
	r.IP = p.String()
}

// holds is true when a is the address the row's port already has, which answers a ping from the node itself and so says nothing about whether it is free
func (r *plannedRow) holds(a netip.Addr) bool {
	return r.port != nil && r.port.IP.IsValid() && r.port.IP.Addr() == a
}

// apply writes what the row sets onto a port, and the tags it adds onto the port's node, and returns what that changed. The MAC is not its business: a row's port either has it already or is being swapped for it
func (r *plannedRow) apply(h *model.Host, p *model.NetInterface) []string {
	var changed []string
	if r.ip.IsValid() && r.ip != p.IP {
		p.IP = r.ip
		changed = append(changed, "changes the address")
	}
	if r.FQDN != "" && r.FQDN != p.FQDN {
		p.FQDN = r.FQDN
		changed = append(changed, "changes the FQDN")
	}
	if r.Ifname != "" && r.Ifname != p.Name {
		p.Name = r.Ifname
		changed = append(changed, "changes the interface name")
	}
	if r.VLAN != "" && r.VLAN != p.VLAN {
		p.VLAN = r.VLAN
		changed = append(changed, "changes the VLAN")
	}
	if r.MTU != 0 && r.MTU != p.MTU {
		p.MTU = r.MTU
		changed = append(changed, "changes the MTU")
	}
	had := len(h.Tags)
	for _, t := range r.Tags {
		if t = strings.TrimSpace(t); t != "" && !slices.Contains(h.Tags, t) {
			h.Tags = append(h.Tags, t)
		}
	}
	if len(h.Tags) > had {
		changed = append(changed, "adds tags")
	}
	return changed
}

type importPlan struct {
	rows []plannedRow
	uses []AddressUse
	// taken is every address the import has spoken for: the ones given, the ones that answered a ping, and each one handed out from a range. The ranges share it, so two that overlap never hand out the same address
	taken map[netip.Addr]bool
}

func (p importPlan) results() []NodeImportResult {
	out := make([]NodeImportResult, len(p.rows))
	for i, r := range p.rows {
		out[i] = r.NodeImportResult
	}
	return out
}

func (p importPlan) summary() NodeImportSummary {
	var s NodeImportSummary
	for _, r := range p.rows {
		switch r.Action {
		case ImportCreate:
			s.Create++
		case ImportAdd:
			s.Add++
		case ImportReplace:
			s.Replace++
		case ImportUnchanged:
			s.Unchanged++
		case ImportConflict:
			s.Conflict++
		}
	}
	return s
}

type macOwner struct {
	node string
	bmc  bool
}

// macOwners maps every MAC on a port in Grendel to the node holding it
func macOwners(hosts model.HostList) map[string]macOwner {
	out := map[string]macOwner{}
	for _, h := range hosts {
		for _, p := range h.Interfaces {
			if len(p.MAC) > 0 {
				out[p.MAC.String()] = macOwner{h.Name, p.BMC}
			}
		}
	}
	return out
}

// planImport decides what each row does to the nodes in hosts, and which address it gets. It changes nothing; hosts builds the nodes to store
func planImport(hosts model.HostList, rows []NodeImportRow, opts importOpts) importPlan {
	p := resolveRows(hosts, rows)
	p.assign(p.allocators(opts))
	return p
}

// planWithPing plans as planImport does, but first pings the addresses it would take from the ranges, with some to spare, and the ones the rows give. It hands out only addresses that were pinged and stayed quiet, so a row is a conflict when too many answered for it to get one
func planWithPing(ctx context.Context, hosts model.HostList, rows []NodeImportRow, opts importOpts) (importPlan, bool, string) {
	p := resolveRows(hosts, rows)
	allocs := p.allocators(opts)

	var toPing []netip.Addr
	for i := range p.rows {
		if r := &p.rows[i]; r.ip.IsValid() && r.Action != ImportUnchanged && !r.holds(r.ip.Addr()) {
			toPing = append(toPing, r.ip.Addr())
		}
	}
	pinged := map[netip.Addr]bool{}
	for _, a := range allocs {
		if n := p.unaddressed(a.bmc); n > 0 {
			for _, addr := range a.candidates(n+max(pingSpare, n/4), pinged) {
				pinged[addr] = true
				toPing = append(toPing, addr)
			}
		}
	}
	if len(toPing) > maxImportPing {
		p.assign(allocs)
		return p, false, fmt.Sprintf("not pinged: more than %d addresses", maxImportPing)
	}

	if len(toPing) > 0 {
		pctx, cancel := context.WithTimeout(ctx, pingTries*pingWait+time.Second)
		answered, err := pingAddrs(pctx, toPing, pingTries, pingWait)
		cancel()
		if err != nil {
			p.assign(allocs)
			return p, false, err.Error()
		}
		for i := range p.rows {
			if r := &p.rows[i]; r.ip.IsValid() && answered[r.ip.Addr()] && !r.holds(r.ip.Addr()) {
				r.Answered = true
			}
		}
		for a := range answered {
			p.taken[a] = true
		}
		for _, a := range allocs {
			a.pinged = pinged
		}
	}
	p.assign(allocs)
	return p, true, ""
}

// resolveRows reads each row, finds the node and port it changes, and marks the rows that conflict with stored nodes or with each other. A row without an address is left without one, for assign to fill in
func resolveRows(hosts model.HostList, rows []NodeImportRow) importPlan {
	byName := map[string]*model.Host{}
	for _, h := range hosts {
		byName[h.Name] = h
	}
	macs := macOwners(hosts)
	uses := addressUses(hosts)
	ipOwner := map[netip.Addr]string{}
	for _, u := range uses {
		ipOwner[u.Addr] = u.Node
	}

	p := importPlan{rows: make([]plannedRow, len(rows)), uses: uses, taken: map[netip.Addr]bool{}}
	macRows := map[string][]*plannedRow{}
	bmcRows := map[string][]*plannedRow{}
	givenIP := map[netip.Addr][]*plannedRow{}

	for i, row := range rows {
		row.Node = strings.TrimSpace(row.Node)
		r := &p.rows[i]
		r.NodeImportRow = row
		hw, err := net.ParseMAC(strings.TrimSpace(row.MAC))
		switch {
		case row.Node == "":
			r.conflict("no node name")
			continue
		case err != nil || len(hw) != 6:
			r.conflict(fmt.Sprintf("%q is not a MAC address", row.MAC))
			continue
		}
		r.mac = hw
		r.MAC = hw.String()
		macRows[r.MAC] = append(macRows[r.MAC], r)
		if row.BMC {
			bmcRows[row.Node] = append(bmcRows[row.Node], r)
		}

		if row.IP != "" {
			ip, err := parseRowIP(row.IP, uses)
			if err != nil {
				r.conflict(fmt.Sprintf("%q is not an IPv4 address", row.IP))
				continue
			}
			r.setIP(ip)
			givenIP[ip.Addr()] = append(givenIP[ip.Addr()], r)
			p.taken[ip.Addr()] = true
		}

		if owner, ok := macs[r.MAC]; ok && owner.node != row.Node {
			r.conflict(fmt.Sprintf("MAC belongs to %s", owner.node))
			continue
		} else if ok && owner.bmc != row.BMC {
			r.conflict(fmt.Sprintf("MAC is already on another port of %s", row.Node))
			continue
		}

		r.host = byName[row.Node]
		if r.host == nil {
			r.Action = ImportCreate
			continue
		}
		r.port = importTarget(r.host, row.BMC, r.mac)
		if r.port == nil {
			r.Action = ImportAdd
			continue
		}
		if !r.ip.IsValid() && r.port.IP.IsValid() {
			r.setIP(r.port.IP)
		}
		if !bytes.Equal(r.port.MAC, r.mac) {
			r.Action = ImportReplace
			r.Replaces = r.port.MAC.String()
			continue
		}
		r.Action = ImportUnchanged
		// applied to copies, as planning only needs to know what the row would change
		node, port := model.Host{Tags: slices.Clone(r.host.Tags)}, *r.port
		if changes := r.apply(&node, &port); len(changes) > 0 {
			r.Action = ImportReplace
			r.Reason = strings.Join(changes, ", ")
		}
	}

	for _, same := range macRows {
		if len(same) > 1 {
			for _, r := range same {
				r.conflict("MAC appears in more than one row")
			}
		}
	}
	for _, same := range bmcRows {
		if len(same) > 1 {
			for _, r := range same {
				r.conflict(fmt.Sprintf("%s has more than one BMC row", r.Node))
			}
		}
	}
	for a, same := range givenIP {
		if len(same) > 1 {
			for _, r := range same {
				r.conflict("address appears in more than one row")
			}
			continue
		}
		if owner, ok := ipOwner[a]; ok && !same[0].holds(a) {
			same[0].conflict(fmt.Sprintf("address is held by %s", owner))
		}
	}
	return p
}

// unaddressed is how many rows for that kind of port still need an address from a range
func (p *importPlan) unaddressed(bmc bool) int {
	n := 0
	for i := range p.rows {
		if r := &p.rows[i]; r.BMC == bmc && r.Action != ImportConflict && !r.ip.IsValid() {
			n++
		}
	}
	return n
}

// allocators is one allocator for each range the import was given, the BMC range first
func (p *importPlan) allocators(opts importOpts) []*allocator {
	var out []*allocator
	for _, kind := range []struct {
		bmc bool
		r   netip.Prefix
	}{{true, opts.bmc}, {false, opts.boot}} {
		if !kind.r.IsValid() {
			continue
		}
		a := &allocator{rangePlan: planRange(kind.r, p.uses, p.taken), bmc: kind.bmc, next: kind.r.Addr()}
		if top := a.highest(); opts.afterHighest && top.IsValid() {
			a.after, a.next = top, top.Next()
		}
		out = append(out, a)
	}
	return out
}

// assign gives every row still without an address the next one its kind of port's range has to hand out
func (p *importPlan) assign(allocs []*allocator) {
	for i := range p.rows {
		r := &p.rows[i]
		if r.Action == ImportConflict || r.ip.IsValid() {
			continue
		}
		at := slices.IndexFunc(allocs, func(a *allocator) bool { return a.bmc == r.BMC })
		if at < 0 {
			kind := "boot"
			if r.BMC {
				kind = "BMC"
			}
			r.conflict(fmt.Sprintf("no address given and no %s range to take one from", kind))
			continue
		}
		addr, why := allocs[at].take()
		if why != "" {
			r.conflict(why)
			continue
		}
		r.setIP(netip.PrefixFrom(addr, allocs[at].bits))
		r.Assigned = true
		if r.Action == ImportUnchanged {
			r.Action = ImportReplace
			r.Reason = "gives the port an address"
		}
	}
}

type allocator struct {
	rangePlan
	bmc bool
	// next is where the search for the next free address resumes
	next netip.Addr
	// after is the highest address taken when the allocator was made, set when it starts past that address rather than at the bottom of the range
	after netip.Addr
	// pinged, once set, holds the only addresses that may be handed out
	pinged map[netip.Addr]bool
	// spent is why nothing more can be handed out, once that is so
	spent string
}

// candidates is the next n free addresses no ping is planned for yet, which are the ones take would hand out first
func (a *allocator) candidates(n int, planned map[netip.Addr]bool) []netip.Addr {
	var out []netip.Addr
	for addr := a.next; len(out) < n && addr.IsValid() && a.r.Contains(addr); addr = addr.Next() {
		if a.free(addr) && !planned[addr] {
			out = append(out, addr)
		}
	}
	return out
}

// take hands out the next free address and marks it taken, or says why there is none
func (a *allocator) take() (netip.Addr, string) {
	if a.spent != "" {
		return netip.Addr{}, a.spent
	}
	for addr := a.next; addr.IsValid() && a.r.Contains(addr); addr = addr.Next() {
		if !a.free(addr) {
			continue
		}
		if a.pinged != nil && !a.pinged[addr] {
			a.spent = fmt.Sprintf("too many addresses in %s answered a ping, and none of those pinged is left", a.r)
			return netip.Addr{}, a.spent
		}
		a.excluded[addr] = true
		a.next = addr.Next()
		return addr, ""
	}
	a.spent = fmt.Sprintf("no free address left in %s", a.r)
	if a.after.IsValid() {
		a.spent = fmt.Sprintf("no free address in %s after %s, the highest in use", a.r, a.after)
	}
	return netip.Addr{}, a.spent
}

// parseRowIP reads a given address. One without a prefix length takes the one its neighbours have
func parseRowIP(s string, uses []AddressUse) (netip.Prefix, error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "/") {
		p, err := netip.ParsePrefix(s)
		if err != nil || !p.Addr().Is4() {
			return netip.Prefix{}, errors.New("not ipv4")
		}
		return p, nil
	}
	a, err := netip.ParseAddr(s)
	if err != nil || !a.Is4() {
		return netip.Prefix{}, errors.New("not ipv4")
	}
	return netip.PrefixFrom(a, subnetBits(netip.PrefixFrom(a, 24).Masked(), uses)), nil
}

// importTarget is the port a row changes: a node's one BMC, whatever its MAC, or the port that is not a BMC with the row's MAC. A row whose MAC is on no such port adds one, so a node can be given more ports than its boot port
func importTarget(h *model.Host, bmc bool, mac net.HardwareAddr) *model.NetInterface {
	if bmc {
		return h.InterfaceBMC()
	}
	if p := h.Interface(mac); p != nil && !p.BMC {
		return p
	}
	return nil
}

// hosts builds every node a commit stores, each a copy of the stored node with the rows' ports applied, since storing a node replaces all of its ports
func (p importPlan) hosts() model.HostList {
	changed := map[string]*model.Host{}
	var out model.HostList
	for i := range p.rows {
		r := &p.rows[i]
		if r.Action == ImportUnchanged || r.Action == ImportConflict {
			continue
		}
		h, ok := changed[r.Node]
		if !ok {
			h = &model.Host{Name: r.Node}
			if r.host != nil {
				h = copyHost(r.host)
			}
			changed[r.Node] = h
			out = append(out, h)
		}

		var port *model.NetInterface
		if r.port != nil {
			// copyHost keeps the ports in order, so the copy's port is at the stored one's place
			port = h.Interfaces[slices.Index(r.host.Interfaces, r.port)]
		} else {
			// a new port goes after the others, so the boot port, the first that is not a BMC, stays the one it was
			port = &model.NetInterface{BMC: r.BMC}
			h.Interfaces = append(h.Interfaces, port)
		}
		port.MAC = r.mac
		r.apply(h, port)
	}
	return out
}

func copyHost(h *model.Host) *model.Host {
	c := *h
	c.Interfaces = make([]*model.NetInterface, len(h.Interfaces))
	for i, p := range h.Interfaces {
		pc := *p
		c.Interfaces[i] = &pc
	}
	c.Bonds = slices.Clone(h.Bonds)
	c.Tags = slices.Clone(h.Tags)
	return &c
}
