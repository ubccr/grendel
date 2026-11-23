// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"testing"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/ubccr/grendel/pkg/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type fakeGNMI struct {
	gpb.UnimplementedGNMIServer
	responses map[string][]*gpb.Notification
	lastReq   *gpb.GetRequest
}

func (f *fakeGNMI) Get(ctx context.Context, req *gpb.GetRequest) (*gpb.GetResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if u := md.Get("username"); len(u) != 1 || u[0] != "admin" {
		return nil, status.Error(codes.Unauthenticated, "bad username")
	}
	f.lastReq = req
	var out []*gpb.Notification
	for _, p := range req.GetPath() {
		out = append(out, f.responses[pathString(p.GetElem())]...)
	}
	return &gpb.GetResponse{Notification: out}, nil
}

func startFake(t *testing.T, f *fakeGNMI, nos string) *GNMI {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	gpb.RegisterGNMIServer(srv, f)
	go srv.Serve(lis)
	t.Cleanup(srv.Stop)

	g := NewGNMI("bufnet", nos, testConfig())
	g.dialOpts = []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	return g
}

// splitPath splits on '/' outside of brackets, since key values such as Ethernet49/1 contain slashes.
func splitPath(p string) []string {
	var parts []string
	depth, start := 0, 0
	for i, r := range p {
		switch r {
		case '[':
			depth++
		case ']':
			depth--
		case '/':
			if depth == 0 {
				parts = append(parts, p[start:i])
				start = i + 1
			}
		}
	}
	return append(parts, p[start:])
}

// parsePath turns an xpath like /a/b[name=x]/c into a gNMI path, so the fixtures can be written the way a switch's paths read. Key values cannot contain ']' or '['.
func parsePath(p string) (*gpb.Path, error) {
	path := &gpb.Path{}
	for _, part := range splitPath(strings.Trim(p, "/")) {
		if part == "" {
			continue
		}
		name, rest, _ := strings.Cut(part, "[")
		elem := &gpb.PathElem{Name: name}
		if rest != "" {
			elem.Key = map[string]string{}
			for _, kv := range strings.Split("["+rest, "[") {
				if kv == "" {
					continue
				}
				kv = strings.TrimSuffix(kv, "]")
				k, v, ok := strings.Cut(kv, "=")
				if !ok {
					return nil, fmt.Errorf("invalid key %q in path %s", kv, p)
				}
				elem.Key[k] = v
			}
		}
		path.Elem = append(path.Elem, elem)
	}
	return path, nil
}

func mustPath(t *testing.T, p string) *gpb.Path {
	t.Helper()
	gp, err := parsePath(p)
	if err != nil {
		t.Fatal(err)
	}
	return gp
}

func byIfname(table model.MACTable) map[string][]*model.MACTableEntry {
	out := map[string][]*model.MACTableEntry{}
	for _, entry := range table {
		out[entry.Ifname] = append(out[entry.Ifname], entry)
	}
	return out
}

func jsonUpdate(t *testing.T, path, js string) *gpb.Update {
	return &gpb.Update{Path: mustPath(t, path), Val: &gpb.TypedValue{Value: &gpb.TypedValue_JsonIetfVal{JsonIetfVal: []byte(js)}}}
}

func strUpdate(t *testing.T, path, v string) *gpb.Update {
	return &gpb.Update{Path: mustPath(t, path), Val: &gpb.TypedValue{Value: &gpb.TypedValue_StringVal{StringVal: v}}}
}

func uintUpdate(t *testing.T, path string, v uint64) *gpb.Update {
	return &gpb.Update{Path: mustPath(t, path), Val: &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{UintVal: v}}}
}

// Arista answers each Get with one JSON_IETF blob per requested path.
func aristaFake(t *testing.T) *fakeGNMI {
	return &fakeGNMI{responses: map[string][]*gpb.Notification{
		"/interfaces/interface/state": {{Update: []*gpb.Update{
			jsonUpdate(t, "/interfaces/interface[name=Ethernet1]/state", `{"openconfig-interfaces:admin-status":"UP","openconfig-interfaces:oper-status":"UP","openconfig-interfaces:mtu":9214,"openconfig-interfaces:description":"cpn-a01-01 bmc"}`),
			jsonUpdate(t, "/interfaces/interface[name=Ethernet2]/state", `{"openconfig-interfaces:admin-status":"UP","openconfig-interfaces:oper-status":"DOWN"}`),
		}}},
		"/network-instances/network-instance[name=default]/fdb/mac-table/entries": {{Update: []*gpb.Update{
			jsonUpdate(t, "/network-instances/network-instance[name=default]/fdb/mac-table/entries", `{"openconfig-network-instance:entry":[
				{"mac-address":"d0:8e:79:00:00:01","vlan":3001,"state":{"mac-address":"d0:8e:79:00:00:01","vlan":3001,"entry-type":"DYNAMIC"},"interface":{"interface-ref":{"state":{"interface":"Ethernet1"}}}},
				{"mac-address":"d0:8e:79:00:00:02","vlan":3001,"state":{"entry-type":"DYNAMIC"},"interface":{"interface-ref":{"state":{"interface":"Ethernet4"}}}},
				{"mac-address":"b8:ca:3a:00:00:02","vlan":3001,"state":{"entry-type":"DYNAMIC"},"interface":{"interface-ref":{"state":{"interface":"Ethernet4"}}}}
			]}`),
		}}},
		"/lldp/interfaces/interface/neighbors/neighbor/state": {{Update: []*gpb.Update{
			jsonUpdate(t, "/lldp/interfaces/interface[name=Ethernet49/1]/neighbors/neighbor[id=1]/state", `{"openconfig-lldp:chassis-id":"7483.ef00.0001","openconfig-lldp:chassis-id-type":"openconfig-lldp-types:MAC_ADDRESS","openconfig-lldp:port-id":"Ethernet1/1","openconfig-lldp:port-id-type":"openconfig-lldp-types:INTERFACE_NAME","openconfig-lldp:system-name":"spine-01.cluster","openconfig-lldp:system-description":"Arista Networks EOS"}`),
		}}},
	}}
}

// SONiC answers with one scalar update per leaf, keys in the path.
func sonicFake(t *testing.T) *fakeGNMI {
	return &fakeGNMI{responses: map[string][]*gpb.Notification{
		"/network-instances/network-instance[name=default]/fdb/mac-table/entries": {{
			Prefix: mustPath(t, "/network-instances/network-instance[name=default]/fdb/mac-table/entries"),
			Update: []*gpb.Update{
				strUpdate(t, "/entry[mac-address=d0:8e:79:00:00:09][vlan=3001]/interface/interface-ref/state/interface", "Ethernet8"),
				uintUpdate(t, "/entry[mac-address=d0:8e:79:00:00:09][vlan=3001]/state/vlan", 3001),
				strUpdate(t, "/entry[mac-address=d0:8e:79:00:00:09][vlan=3001]/state/entry-type", "DYNAMIC"),
				strUpdate(t, "/entry[mac-address=d0:8e:79:00:00:09][vlan=1]/interface/interface-ref/state/interface", "Ethernet12"),
			},
		}},
		"/lldp/interfaces/interface/neighbors/neighbor/state": {{Update: []*gpb.Update{
			strUpdate(t, "/lldp/interfaces/interface[name=Ethernet0]/neighbors/neighbor[id=Ethernet0]/state/system-name", "cpn-a01-01"),
			strUpdate(t, "/lldp/interfaces/interface[name=Ethernet0]/neighbors/neighbor[id=Ethernet0]/state/port-id", "b8:ca:3a:00:00:01"),
			strUpdate(t, "/lldp/interfaces/interface[name=Ethernet0]/neighbors/neighbor[id=Ethernet0]/state/port-id-type", "MAC_ADDRESS"),
		}}},
	}}
}

func macsOn(t *testing.T, g *GNMI) map[string][]string {
	t.Helper()
	table, err := g.GetMACTable(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]string{}
	for ifname, entries := range byIfname(table) {
		for _, e := range entries {
			out[ifname] = append(out[ifname], e.MAC.String()+"@"+e.VLAN)
		}
		sort.Strings(out[ifname])
	}
	return out
}

func TestGNMIAristaMACTable(t *testing.T) {
	f := aristaFake(t)
	got := macsOn(t, startFake(t, f, NOSArista))

	want := map[string][]string{
		"Ethernet1": {"d0:8e:79:00:00:01@3001"},
		"Ethernet4": {"b8:ca:3a:00:00:02@3001", "d0:8e:79:00:00:02@3001"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for k, v := range want {
		if len(got[k]) != len(v) || got[k][0] != v[0] {
			t.Errorf("%s = %v, want %v", k, got[k], v)
		}
	}

	if f.lastReq.GetEncoding() != gpb.Encoding_JSON_IETF || f.lastReq.GetType() != gpb.GetRequest_STATE {
		t.Errorf("request = %v, want JSON_IETF STATE", f.lastReq)
	}
}

func TestGNMISonicMACTableFromLeaves(t *testing.T) {
	got := macsOn(t, startFake(t, sonicFake(t), NOSSonic))
	if len(got["Ethernet8"]) != 1 || got["Ethernet8"][0] != "d0:8e:79:00:00:09@3001" {
		t.Errorf("Ethernet8 = %v", got["Ethernet8"])
	}
	if len(got["Ethernet12"]) != 1 || got["Ethernet12"][0] != "d0:8e:79:00:00:09@1" {
		t.Errorf("the same mac in another vlan must be its own entry: %v", got)
	}
}

func TestGNMIInterfaces(t *testing.T) {
	got, err := startFake(t, aristaFake(t), NOSArista).GetInterfaceDescriptions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got["Ethernet1"] != "cpn-a01-01 bmc" {
		t.Errorf("Ethernet1 = %q", got["Ethernet1"])
	}
	if d, ok := got["Ethernet2"]; !ok || d != "" {
		t.Errorf("an interface with no description must still be listed: %v", got)
	}
}

func TestGNMIFreeTextKeepsItsColons(t *testing.T) {
	leaves := func(updates ...*gpb.Update) []leaf {
		out, err := leavesFromNotifications([]*gpb.Notification{{Update: updates}})
		if err != nil {
			t.Fatal(err)
		}
		return out
	}

	descriptions := parseInterfaces(leaves(
		jsonUpdate(t, "/interfaces/interface[name=Ethernet7]/state", `{"openconfig-interfaces:description":"cpn-a01-07:eth0"}`),
		strUpdate(t, "/interfaces/interface[name=Ethernet8]/state/description", "cpn-a01-08:eth0"),
	))
	if descriptions["Ethernet7"] != "cpn-a01-07:eth0" || descriptions["Ethernet8"] != "cpn-a01-08:eth0" {
		t.Errorf("descriptions = %v, want them as the switch sent them", descriptions)
	}

	n := parseLLDP(leaves(
		jsonUpdate(t, "/lldp/interfaces/interface[name=Ethernet7]/neighbors/neighbor[id=1]/state", `{"openconfig-lldp:system-name":"cpn-a01-07:idrac","openconfig-lldp:port-id-type":"openconfig-lldp-types:MAC_ADDRESS"}`),
	))["Ethernet7"]
	if n == nil || n.SystemName != "cpn-a01-07:idrac" || n.PortIdType != "MAC_ADDRESS" {
		t.Errorf("neighbor = %+v, want the system name whole and only the identityref stripped", n)
	}
}

func TestGNMIRequestPaths(t *testing.T) {
	f := aristaFake(t)
	g := startFake(t, f, NOSArista)
	g.cfg.NetworkInstance = "Vlan3001"
	g.GetMACTable(context.Background())
	if got := pathString(f.lastReq.GetPath()[0].GetElem()); got != "/network-instances/network-instance[name=Vlan3001]/fdb/mac-table/entries" {
		t.Errorf("mac table path = %s", got)
	}
}

func TestDecodeValueScalars(t *testing.T) {
	for want, v := range map[string]*gpb.TypedValue{
		"1.5":   {Value: &gpb.TypedValue_DoubleVal{DoubleVal: 1.5}},
		"-3":    {Value: &gpb.TypedValue_IntVal{IntVal: -3}},
		"true":  {Value: &gpb.TypedValue_BoolVal{BoolVal: true}},
		"ascii": {Value: &gpb.TypedValue_AsciiVal{AsciiVal: "ascii"}},
		"3001":  {Value: &gpb.TypedValue_JsonIetfVal{JsonIetfVal: []byte(`3001`)}},
		"":      {Value: &gpb.TypedValue_AnyVal{}},
	} {
		got, _, err := decodeValue(v)
		if err != nil {
			t.Fatal(err)
		}
		if s := (leaf{value: got}).str(); s != want {
			t.Errorf("decodeValue(%v) reads as %q, want %q", v, s, want)
		}
	}
}

func TestGNMILLDP(t *testing.T) {
	n, err := startFake(t, aristaFake(t), NOSArista).GetLLDPNeighbors(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	up := n["Ethernet49/1"]
	if up == nil {
		t.Fatalf("no neighbor on Ethernet49/1: %v", n)
	}
	if up.ChassisIdType != "MAC_ADDRESS" || up.PortIdType != "INTERFACE_NAME" || up.SystemName != "spine-01.cluster" {
		t.Errorf("module prefixes not stripped or fields missing: %+v", up)
	}
	if up.ChassisId != "74:83:ef:00:00:01" {
		t.Errorf("chassis id = %q, want it normalised to colon form", up.ChassisId)
	}

	n, err = startFake(t, sonicFake(t), NOSSonic).GetLLDPNeighbors(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if e0 := n["Ethernet0"]; e0 == nil || e0.SystemName != "cpn-a01-01" || e0.PortIdType != "MAC_ADDRESS" {
		t.Errorf("Ethernet0 = %+v", e0)
	}
}

func TestGNMIRejectsBadCredentials(t *testing.T) {
	g := startFake(t, aristaFake(t), NOSArista)
	g.username = "nobody"
	if _, err := g.GetMACTable(context.Background()); status.Code(err) != codes.Unauthenticated {
		t.Errorf("err = %v, want Unauthenticated", err)
	}
}

func TestParsePath(t *testing.T) {
	p := mustPath(t, "/network-instances/network-instance[name=default]/fdb/mac-table/entries/entry[mac-address=aa:bb:cc:dd:ee:ff][vlan=10]")
	if got := pathString(p.GetElem()); got != "/network-instances/network-instance[name=default]/fdb/mac-table/entries/entry[mac-address=aa:bb:cc:dd:ee:ff][vlan=10]" {
		t.Errorf("round trip = %s", got)
	}
	if _, err := parsePath("/a/b[novalue]"); err == nil {
		t.Error("expected an error for a key without a value")
	}
}

func TestStripModule(t *testing.T) {
	for in, want := range map[string]string{
		"openconfig-lldp-types:MAC_ADDRESS": "MAC_ADDRESS",
		"openconfig-interfaces:mtu":         "mtu",
		"MAC_ADDRESS":                       "MAC_ADDRESS",
		"mac-table":                         "mac-table",
	} {
		if got := stripModule(in); got != want {
			t.Errorf("stripModule(%q) = %q, want %q", in, got, want)
		}
	}
}
