// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
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
	// importPingRounds bounds how many times addresses that answered a ping are swapped for the next free ones
	importPingRounds = 3
	maxImportPing    = 2048
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
	res.Rows, res.Summary = plan.rows, plan.summary()

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
	bmc, boot      netip.Prefix
	afterHighest   bool
	excludeAnswers map[netip.Addr]bool
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

// planWithPing plans, pings the addresses it took from the ranges, and plans again without the ones that answered, a few rounds at most
func planWithPing(ctx context.Context, hosts model.HostList, rows []NodeImportRow, opts importOpts) (importPlan, bool, string) {
	opts.excludeAnswers = map[netip.Addr]bool{}
	asked := map[netip.Addr]bool{}
	var plan importPlan
	for range importPingRounds {
		plan = planImport(hosts, rows, opts)
		var toPing []netip.Addr
		for _, r := range plan.rows {
			if r.IP == "" {
				continue
			}
			a := netip.MustParsePrefix(r.IP).Addr()
			// the address a port already has answers from the node itself, which says nothing about whether it is free
			if !asked[a] && (r.Assigned || r.Action != ImportUnchanged) && !ownPort(plan.existing[r.Node], r.BMC, r.MAC, a) {
				toPing = append(toPing, a)
			}
		}
		if len(toPing) == 0 {
			break
		}
		if len(asked)+len(toPing) > maxImportPing {
			return plan, false, fmt.Sprintf("not pinged: more than %d addresses", maxImportPing)
		}
		for _, a := range toPing {
			asked[a] = true
		}

		pctx, cancel := context.WithTimeout(ctx, pingTries*pingWait+time.Second)
		answered, err := pingAddrs(pctx, toPing, pingTries, pingWait)
		cancel()
		if err != nil {
			return plan, false, err.Error()
		}
		fresh := false
		for a := range answered {
			if !opts.excludeAnswers[a] {
				opts.excludeAnswers[a] = true
				fresh = true
			}
		}
		if !fresh {
			break
		}
	}
	plan = planImport(hosts, rows, opts)
	for i, r := range plan.rows {
		if r.IP == "" || r.Assigned {
			continue
		}
		if a := netip.MustParsePrefix(r.IP).Addr(); opts.excludeAnswers[a] && !ownPort(plan.existing[r.Node], r.BMC, r.MAC, a) {
			plan.rows[i].Answered = true
		}
	}
	return plan, true, ""
}

type importPlan struct {
	rows     []NodeImportResult
	existing map[string]*model.Host
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

	out := make([]NodeImportResult, len(rows))
	macRows := map[string][]int{}
	portRows := map[string][]int{}
	givenIP := map[netip.Addr][]int{}

	conflict := func(i int, reason string) {
		if out[i].Action != ImportConflict {
			out[i].Action = ImportConflict
			out[i].Reason = reason
		}
	}

	for i, row := range rows {
		row.Node = strings.TrimSpace(row.Node)
		out[i] = NodeImportResult{NodeImportRow: row}
		hw, err := net.ParseMAC(strings.TrimSpace(row.MAC))
		switch {
		case row.Node == "":
			conflict(i, "no node name")
			continue
		case err != nil || len(hw) != 6:
			conflict(i, fmt.Sprintf("%q is not a MAC address", row.MAC))
			continue
		}
		mac := hw.String()
		out[i].MAC = mac
		macRows[mac] = append(macRows[mac], i)
		if row.BMC {
			portRows[row.Node] = append(portRows[row.Node], i)
		}

		if row.IP != "" {
			p, err := parseRowIP(row.IP, uses)
			if err != nil {
				conflict(i, fmt.Sprintf("%q is not an IPv4 address", row.IP))
				continue
			}
			out[i].IP = p.String()
			givenIP[p.Addr()] = append(givenIP[p.Addr()], i)
		}

		if owner, ok := macs[mac]; ok && owner.node != row.Node {
			conflict(i, fmt.Sprintf("MAC belongs to %s", owner.node))
			continue
		} else if ok && owner.bmc != row.BMC {
			conflict(i, fmt.Sprintf("MAC is already on another port of %s", row.Node))
			continue
		}

		host := byName[row.Node]
		if host == nil {
			out[i].Action = ImportCreate
			continue
		}
		port := importTarget(host, row.BMC, mac)
		switch {
		case port == nil:
			out[i].Action = ImportAdd
		case port.MAC.String() == mac:
			if out[i].IP == "" && port.IP.IsValid() {
				out[i].IP = port.IP.String()
			}
			out[i].Action = ImportUnchanged
			if changes := portChanges(host, port, out[i].NodeImportRow); len(changes) > 0 {
				out[i].Action = ImportReplace
				out[i].Reason = strings.Join(changes, ", ")
			}
		default:
			out[i].Action = ImportReplace
			out[i].Replaces = port.MAC.String()
			if out[i].IP == "" && port.IP.IsValid() {
				out[i].IP = port.IP.String()
			}
		}
	}

	for _, idx := range macRows {
		if len(idx) > 1 {
			for _, i := range idx {
				conflict(i, "MAC appears in more than one row")
			}
		}
	}
	for _, idx := range portRows {
		if len(idx) > 1 {
			for _, i := range idx {
				conflict(i, fmt.Sprintf("%s has more than one BMC row", out[i].Node))
			}
		}
	}
	for a, idx := range givenIP {
		if len(idx) > 1 {
			for _, i := range idx {
				conflict(i, "address appears in more than one row")
			}
			continue
		}
		i := idx[0]
		if owner, ok := ipOwner[a]; ok && !ownPort(byName[out[i].Node], out[i].BMC, out[i].MAC, a) {
			conflict(i, fmt.Sprintf("address is held by %s", owner))
		}
	}

	var exclude []netip.Addr
	for a := range givenIP {
		exclude = append(exclude, a)
	}
	for a := range opts.excludeAnswers {
		exclude = append(exclude, a)
	}
	ranges := map[bool]*rangePlan{}
	for bmc, r := range map[bool]netip.Prefix{true: opts.bmc, false: opts.boot} {
		if r.IsValid() {
			p := planRange(r, uses, exclude)
			ranges[bmc] = &p
		}
	}
	next := map[bool]netip.Addr{}
	for i := range out {
		r := &out[i]
		if r.Action == ImportConflict || r.IP != "" {
			continue
		}
		p := ranges[r.BMC]
		if p == nil {
			kind := "boot"
			if r.BMC {
				kind = "BMC"
			}
			conflict(i, fmt.Sprintf("no address given and no %s range to take one from", kind))
			continue
		}
		start, ok := next[r.BMC]
		if !ok {
			start = p.r.Addr()
			if opts.afterHighest {
				if top := p.highest(); top.IsValid() {
					start = top.Next()
				}
			}
		}
		free := p.freeFrom(start, 1)
		if len(free) == 0 {
			conflict(i, fmt.Sprintf("no free address left in %s", p.r))
			continue
		}
		r.IP = netip.PrefixFrom(free[0], p.bits).String()
		r.Assigned = true
		next[r.BMC] = free[0].Next()
		if r.Action == ImportUnchanged {
			r.Action = ImportReplace
			r.Reason = "gives the port an address"
		}
	}

	return importPlan{rows: out, existing: byName}
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
func importTarget(h *model.Host, bmc bool, mac string) *model.NetInterface {
	if bmc {
		return h.InterfaceBMC()
	}
	for _, p := range h.Interfaces {
		if !p.BMC && p.MAC.String() == mac {
			return p
		}
	}
	return nil
}

// portChanges is what a row would change on a port that already has its MAC: every field the row sets that differs, and tags the node lacks
func portChanges(h *model.Host, p *model.NetInterface, row NodeImportRow) []string {
	var out []string
	if row.IP != "" && row.IP != p.IP.String() {
		out = append(out, "changes the address")
	}
	if row.FQDN != "" && row.FQDN != p.FQDN {
		out = append(out, "changes the FQDN")
	}
	if row.Ifname != "" && row.Ifname != p.Name {
		out = append(out, "changes the interface name")
	}
	if row.VLAN != "" && row.VLAN != p.VLAN {
		out = append(out, "changes the VLAN")
	}
	if row.MTU != 0 && row.MTU != p.MTU {
		out = append(out, "changes the MTU")
	}
	for _, t := range row.Tags {
		if t = strings.TrimSpace(t); t != "" && !slices.Contains(h.Tags, t) {
			out = append(out, "adds tags")
			break
		}
	}
	return out
}

func ownPort(h *model.Host, bmc bool, mac string, a netip.Addr) bool {
	if h == nil {
		return false
	}
	p := importTarget(h, bmc, mac)
	return p != nil && p.IP.IsValid() && p.IP.Addr() == a
}

// hosts builds every node a commit stores, each a copy of the stored node with the rows' ports applied, since storing a node replaces all of its ports
func (p importPlan) hosts() model.HostList {
	changed := map[string]*model.Host{}
	var order []string
	for _, r := range p.rows {
		if r.Action == ImportUnchanged || r.Action == ImportConflict {
			continue
		}
		h, ok := changed[r.Node]
		if !ok {
			if old := p.existing[r.Node]; old != nil {
				h = copyHost(old)
			} else {
				h = &model.Host{Name: r.Node}
			}
			changed[r.Node] = h
			order = append(order, r.Node)
		}

		mac, _ := net.ParseMAC(r.MAC)
		ip, _ := netip.ParsePrefix(r.IP)
		port := importTarget(h, r.BMC, r.MAC)
		if r.Action == ImportCreate || r.Action == ImportAdd || port == nil {
			// a new port goes after the others, so the boot port, the first that is not a BMC, stays the one it was
			port = &model.NetInterface{BMC: r.BMC}
			h.Interfaces = append(h.Interfaces, port)
		}
		port.MAC = mac
		port.IP = ip
		if r.FQDN != "" {
			port.FQDN = r.FQDN
		}
		if r.Ifname != "" {
			port.Name = r.Ifname
		}
		if r.VLAN != "" {
			port.VLAN = r.VLAN
		}
		if r.MTU != 0 {
			port.MTU = r.MTU
		}
		for _, t := range r.Tags {
			if t = strings.TrimSpace(t); t != "" && !slices.Contains(h.Tags, t) {
				h.Tags = append(h.Tags, t)
			}
		}
	}

	out := make(model.HostList, 0, len(order))
	for _, n := range order {
		out = append(out, changed[n])
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
