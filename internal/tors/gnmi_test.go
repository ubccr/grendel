// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"context"
	"net"
	"sort"
	"testing"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
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

func mustPath(t *testing.T, p string) *gpb.Path {
	t.Helper()
	gp, err := parsePath(p)
	if err != nil {
		t.Fatal(err)
	}
	return gp
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
	for ifname, entries := range table.ByIfname() {
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
	table, err := startFake(t, aristaFake(t), NOSArista).GetInterfaceStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	e1 := table["Ethernet1"]
	if e1 == nil || e1.LineProtocolStatus != "up" || e1.MTU != 9214 || e1.Description != "cpn-a01-01 bmc" {
		t.Errorf("Ethernet1 = %+v", e1)
	}
	if e2 := table["Ethernet2"]; e2 == nil || e2.LineProtocolStatus != "down" {
		t.Errorf("Ethernet2 = %+v", e2)
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
		"d0:8e:79:00:00:01":                 "d0:8e:79:00:00:01",
		"Ethernet1/1":                       "Ethernet1/1",
		"10.0.0.1":                          "10.0.0.1",
		"fe80::1":                           "fe80::1",
	} {
		if got := stripModule(in); got != want {
			t.Errorf("stripModule(%q) = %q, want %q", in, got, want)
		}
	}
}
