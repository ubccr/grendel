// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"cmp"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/ubccr/grendel/pkg/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GNMI struct {
	addr     string
	nos      string
	cfg      GNMIConfig
	username string
	password string

	dialOpts []grpc.DialOption

	mu   sync.Mutex
	conn *grpc.ClientConn
}

func NewGNMI(addr, nos string, cfg Config) *GNMI {
	return &GNMI{
		addr:     addr,
		nos:      nos,
		cfg:      cfg.gnmiFor(nos),
		username: cfg.Username,
		password: cfg.Password,
	}
}

func (g *GNMI) target() string {
	return net.JoinHostPort(g.addr, strconv.Itoa(g.cfg.Port))
}

func (g *GNMI) transport() (credentials.TransportCredentials, error) {
	if g.cfg.Plaintext {
		return insecure.NewCredentials(), nil
	}
	tc := &tls.Config{InsecureSkipVerify: g.cfg.SkipVerify && g.cfg.CACert == ""}
	if g.cfg.CACert != "" {
		pem, err := os.ReadFile(g.cfg.CACert)
		if err != nil {
			return nil, fmt.Errorf("reading switch.gnmi.ca_cert: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("no certificates found in %s", g.cfg.CACert)
		}
		tc.RootCAs = pool
	}
	return credentials.NewTLS(tc), nil
}

// client opens the connection on first use and shares it, so the calls of one scan make one TLS handshake between them
func (g *GNMI) client() (*grpc.ClientConn, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.conn != nil {
		return g.conn, nil
	}
	opts := g.dialOpts
	if opts == nil {
		creds, err := g.transport()
		if err != nil {
			return nil, err
		}
		opts = []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	}
	conn, err := grpc.NewClient("passthrough:///"+g.target(), opts...)
	if err != nil {
		return nil, err
	}
	g.conn = conn
	return conn, nil
}

func (g *GNMI) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.conn == nil {
		return nil
	}
	err := g.conn.Close()
	g.conn = nil
	return err
}

func (g *GNMI) get(ctx context.Context, path *gpb.Path) ([]leaf, error) {
	conn, err := g.client()
	if err != nil {
		return nil, err
	}

	req := &gpb.GetRequest{Type: gpb.GetRequest_STATE, Encoding: gpb.Encoding_JSON_IETF, Path: []*gpb.Path{path}}
	if g.cfg.Target != "" {
		req.Prefix = &gpb.Path{Target: g.cfg.Target}
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "username", g.username, "password", g.password)
	// a switch answers with its whole MAC table in one message, which a full table takes past grpc's 4 MB default
	res, err := gpb.NewGNMIClient(conn).Get(ctx, req, grpc.MaxCallRecvMsgSize(math.MaxInt32))
	if err != nil {
		return nil, fmt.Errorf("gnmi get from %s: %w", g.target(), err)
	}
	return leavesFromNotifications(res.GetNotification())
}

type record struct {
	keys   map[string]map[string]string
	fields map[string]string
}

func (r record) key(elem, k string) string {
	return r.keys[elem][k]
}

// ident reads a field holding an identityref, which JSON_IETF prefixes with its module
func (r record) ident(field string) string {
	return stripModule(r.fields[field])
}

// group gathers leaves into one record per instance of the list called elem. Each record holds the keys of every path element down to that list and the leaves below it, named by their path relative to the list item, e.g. state/vlan.
func group(leaves []leaf, elem string, skip ...string) []record {
	byPath := map[string]*record{}
	var order []string
next:
	for _, l := range leaves {
		at := -1
		for i, e := range l.elems {
			for _, s := range skip {
				if e.Name == s {
					continue next
				}
			}
			if e.Name == elem && at < 0 {
				at = i
			}
		}
		if at < 0 || at == len(l.elems)-1 {
			continue
		}
		id := pathString(l.elems[:at+1])
		r, ok := byPath[id]
		if !ok {
			r = &record{keys: map[string]map[string]string{}, fields: map[string]string{}}
			for _, e := range l.elems[:at+1] {
				if len(e.Key) > 0 {
					r.keys[e.Name] = e.Key
				}
			}
			byPath[id] = r
			order = append(order, id)
		}
		names := make([]string, 0, len(l.elems)-at-1)
		for _, e := range l.elems[at+1:] {
			names = append(names, e.Name)
		}
		r.fields[strings.Join(names, "/")] = l.str()
	}

	out := make([]record, 0, len(order))
	for _, id := range order {
		out = append(out, *byPath[id])
	}
	return out
}

// GetInterfaceDescriptions asks for the whole state container rather than its description leaf, since an interface with no description has no such leaf and would be missing from the answer
func (g *GNMI) GetInterfaceDescriptions(ctx context.Context) (map[string]string, error) {
	leaves, err := g.get(ctx, elems("interfaces", "interface", "state"))
	if err != nil {
		return nil, err
	}
	return parseInterfaces(leaves), nil
}

func parseInterfaces(leaves []leaf) map[string]string {
	out := map[string]string{}
	for _, r := range group(leaves, "interface", "subinterfaces") {
		name := cmp.Or(r.key("interface", "name"), r.fields["state/name"], r.fields["name"])
		if name == "" {
			continue
		}
		out[name] = r.fields["state/description"]
	}
	return out
}

func (g *GNMI) GetMACTable(ctx context.Context) (model.MACTable, error) {
	path := elems("network-instances", "network-instance", "fdb", "mac-table", "entries")
	path.Elem[1].Key = map[string]string{"name": g.cfg.NetworkInstance}
	leaves, err := g.get(ctx, path)
	if err != nil {
		return nil, err
	}
	return parseFDB(leaves), nil
}

func parseFDB(leaves []leaf) model.MACTable {
	out := model.MACTable{}
	for _, r := range group(leaves, "entry") {
		macStr := cmp.Or(r.key("entry", "mac-address"), r.fields["state/mac-address"], r.fields["mac-address"])
		mac, err := net.ParseMAC(macStr)
		if err != nil {
			log.Debugf("skipping fdb entry with unreadable mac %q", macStr)
			continue
		}
		ifname := cmp.Or(r.fields["interface/interface-ref/state/interface"], r.fields["interface/interface-ref/config/interface"])
		if ifname == "" {
			continue
		}
		out = append(out, &model.MACTableEntry{
			Ifname: ifname,
			VLAN:   cmp.Or(r.key("entry", "vlan"), r.fields["state/vlan"], r.fields["vlan"]),
			Type:   strings.ToLower(r.ident("state/entry-type")),
			MAC:    mac,
		})
	}
	return out
}

func (g *GNMI) GetLLDPNeighbors(ctx context.Context) (model.LLDPNeighbors, error) {
	leaves, err := g.get(ctx, elems("lldp", "interfaces", "interface", "neighbors", "neighbor", "state"))
	if err != nil {
		return nil, err
	}
	return parseLLDP(leaves), nil
}

func parseLLDP(leaves []leaf) model.LLDPNeighbors {
	out := model.LLDPNeighbors{}
	for _, r := range group(leaves, "neighbor") {
		port := r.key("interface", "name")
		if port == "" {
			continue
		}
		chassisType := r.ident("state/chassis-id-type")
		chassis := r.fields["state/chassis-id"]
		if chassisType == "MAC_ADDRESS" {
			if mac, err := net.ParseMAC(chassis); err == nil {
				chassis = mac.String()
			}
		}
		out[port] = &model.LLDP{
			PortName:          port,
			ChassisIdType:     chassisType,
			ChassisId:         chassis,
			SystemName:        r.fields["state/system-name"],
			SystemDescription: r.fields["state/system-description"],
			ManagementAddress: r.fields["state/management-address"],
			PortDescription:   r.fields["state/port-description"],
			PortId:            r.fields["state/port-id"],
			PortIdType:        r.ident("state/port-id-type"),
		}
	}
	return out
}
