// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/tors"
	"github.com/ubccr/grendel/pkg/model"
)

func port(mac, ip string, bmc bool) *model.NetInterface {
	hw, _ := net.ParseMAC(mac)
	p := &model.NetInterface{MAC: hw, BMC: bmc}
	if ip != "" {
		p.IP = netip.MustParsePrefix(ip)
	}
	return p
}

func importHosts() model.HostList {
	return model.HostList{
		{Name: "cpn-a01-01", Interfaces: []*model.NetInterface{
			port("b8:ca:3a:00:00:01", "10.2.0.1/22", false),
			port("d0:8e:79:00:00:01", "10.1.0.1/22", true),
		}},
		{Name: "cpn-a01-02", Interfaces: []*model.NetInterface{
			port("b8:ca:3a:00:00:02", "10.2.0.2/22", false),
		}},
		{Name: "cpn-a01-03", Interfaces: []*model.NetInterface{
			port("d0:8e:79:00:00:03", "10.1.0.3/22", true),
		}},
	}
}

func bmcOpts() importOpts {
	return importOpts{bmc: netip.MustParsePrefix("10.1.0.0/22"), boot: netip.MustParsePrefix("10.2.0.0/22")}
}

func actions(p importPlan) []string {
	out := []string{}
	for _, r := range p.rows {
		out = append(out, r.Action)
	}
	return out
}

func TestPlanImportActions(t *testing.T) {
	rows := []NodeImportRow{
		{Node: "cpn-a01-01", MAC: "D0-8E-79-00-00-01", BMC: true},
		{Node: "cpn-a01-02", MAC: "d0:8e:79:00:00:02", BMC: true},
		{Node: "cpn-a01-03", MAC: "d0:8e:79:00:00:33", BMC: true},
		{Node: "cpn-a01-04", MAC: "d0:8e:79:00:00:04", BMC: true},
	}
	p := planImport(importHosts(), rows, bmcOpts())

	assert.Equal(t, []string{ImportUnchanged, ImportAdd, ImportReplace, ImportCreate}, actions(p))
	assert.Equal(t, "d0:8e:79:00:00:01", p.rows[0].MAC, "MACs are normalised")
	assert.Equal(t, "10.1.0.1/22", p.rows[0].IP, "an unchanged port keeps its address")
	assert.False(t, p.rows[0].Assigned)
	assert.Equal(t, "d0:8e:79:00:00:03", p.rows[2].Replaces)
	assert.Equal(t, "10.1.0.3/22", p.rows[2].IP, "a replaced BMC keeps its address")

	assert.Equal(t, "10.1.0.2/22", p.rows[1].IP, "the lowest free address, with the prefix of its neighbours")
	assert.Equal(t, "10.1.0.4/22", p.rows[3].IP, "assigned in row order, skipping ones in use")
	assert.True(t, p.rows[3].Assigned)
}

func TestPlanImportUnchangedPortFields(t *testing.T) {
	row := NodeImportRow{Node: "cpn-a01-01", MAC: "d0:8e:79:00:00:01", BMC: true}
	cases := map[string]struct {
		edit   func(*NodeImportRow)
		reason string
	}{
		"vlan":   {func(r *NodeImportRow) { r.VLAN = "3001" }, "changes the VLAN"},
		"mtu":    {func(r *NodeImportRow) { r.MTU = 9000 }, "changes the MTU"},
		"ifname": {func(r *NodeImportRow) { r.Ifname = "bmc0" }, "changes the interface name"},
		"tags":   {func(r *NodeImportRow) { r.Tags = []string{"rack:a01"} }, "adds tags"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			r := row
			c.edit(&r)
			p := planImport(importHosts(), []NodeImportRow{r}, bmcOpts())
			assert.Equal(t, ImportReplace, p.rows[0].Action)
			assert.Equal(t, c.reason, p.rows[0].Reason)

			h := p.hosts()
			require.Len(t, h, 1)
			bmc := h[0].InterfaceBMC()
			assert.Equal(t, "d0:8e:79:00:00:01", bmc.MAC.String())
			assert.Equal(t, "10.1.0.1/22", bmc.IP.String(), "the address is kept")
			if name == "tags" {
				assert.Contains(t, h[0].Tags, "rack:a01")
			}
		})
	}
}

func TestPlanWithPingSkipsOwnAddress(t *testing.T) {
	asked := fakePing(t, "10.1.0.1")
	rows := []NodeImportRow{{Node: "cpn-a01-01", MAC: "d0:8e:79:00:00:01", BMC: true, FQDN: "bmc-cpn-a01-01.cluster"}}
	p, pinged, perr := planWithPing(context.Background(), importHosts(), rows, bmcOpts())
	require.True(t, pinged, perr)
	assert.Equal(t, ImportReplace, p.rows[0].Action)
	assert.NotContains(t, *asked, netip.MustParseAddr("10.1.0.1"), "the node's own address answers from the node")
	assert.False(t, p.rows[0].Answered)
}

func TestPlanImportConflicts(t *testing.T) {
	rows := []NodeImportRow{
		{Node: "cpn-a01-09", MAC: "d0:8e:79:00:00:01", BMC: true},
		{Node: "cpn-a01-10", MAC: "aa:aa:aa:00:00:01", BMC: true},
		{Node: "cpn-a01-11", MAC: "aa:aa:aa:00:00:01", BMC: true},
		{Node: "cpn-a01-12", MAC: "aa:aa:aa:00:00:12", BMC: true},
		{Node: "cpn-a01-12", MAC: "aa:aa:aa:00:00:13", BMC: true},
		{Node: "cpn-a01-14", MAC: "not-a-mac", BMC: true},
		{Node: "cpn-a01-15", MAC: "aa:aa:aa:00:00:15", BMC: true, IP: "10.1.0.3"},
		{Node: "cpn-a01-16", MAC: "aa:aa:aa:00:00:16", BMC: true, IP: "10.1.0.200"},
		{Node: "cpn-a01-17", MAC: "aa:aa:aa:00:00:17", BMC: true, IP: "10.1.0.200"},
		{Node: "", MAC: "aa:aa:aa:00:00:18", BMC: true},
		{Node: "cpn-a01-02", MAC: "b8:ca:3a:00:00:02", BMC: true},
	}
	p := planImport(importHosts(), rows, bmcOpts())

	reasons := map[string]string{}
	for _, r := range p.rows {
		require.Equal(t, ImportConflict, r.Action, "row %+v", r.NodeImportRow)
		reasons[r.Node+" "+r.MAC] = r.Reason
	}
	assert.Equal(t, "MAC belongs to cpn-a01-01", reasons["cpn-a01-09 d0:8e:79:00:00:01"])
	assert.Equal(t, "MAC appears in more than one row", reasons["cpn-a01-10 aa:aa:aa:00:00:01"])
	assert.Equal(t, "cpn-a01-12 has more than one BMC row", reasons["cpn-a01-12 aa:aa:aa:00:00:12"])
	assert.Contains(t, reasons["cpn-a01-14 not-a-mac"], "is not a MAC address")
	assert.Equal(t, "address is held by cpn-a01-03", reasons["cpn-a01-15 aa:aa:aa:00:00:15"])
	assert.Equal(t, "address appears in more than one row", reasons["cpn-a01-16 aa:aa:aa:00:00:16"])
	assert.Equal(t, "no node name", reasons[" aa:aa:aa:00:00:18"])
	assert.Equal(t, "MAC is already on another port of cpn-a01-02", reasons["cpn-a01-02 b8:ca:3a:00:00:02"])
	assert.Equal(t, len(rows), p.summary().Conflict)
}

func TestPlanImportBootRows(t *testing.T) {
	rows := []NodeImportRow{
		{Node: "cpn-a01-03", MAC: "b8:ca:3a:00:00:03", FQDN: "cpn-a01-03.cluster", Ifname: "eno1"},
		{Node: "cpn-a01-01", MAC: "b8:ca:3a:00:00:91"},
		{Node: "cpn-a01-01", MAC: "b8:ca:3a:00:00:92"},
		{Node: "cpn-a01-02", MAC: "b8:ca:3a:00:00:02", FQDN: "cpn-a01-02.cluster"},
	}
	p := planImport(importHosts(), rows, bmcOpts())
	assert.Equal(t, []string{ImportAdd, ImportAdd, ImportAdd, ImportReplace}, actions(p), "a node can be given several ports, and one already there is updated")
	assert.Empty(t, p.rows[3].Replaces, "an update keeps the port's MAC")
	assert.Equal(t, "10.2.0.3/22", p.rows[0].IP)
	assert.Equal(t, "10.2.0.4/22", p.rows[1].IP)
	assert.Equal(t, "10.2.0.2/22", p.rows[3].IP, "an updated port keeps its address")

	hosts := p.hosts()
	require.Len(t, hosts, 3)
	n3 := hosts[0]
	require.Len(t, n3.Interfaces, 2)
	assert.Equal(t, "b8:ca:3a:00:00:03", n3.BootInterface().MAC.String(), "a node's first port that isn't its BMC is its boot port")
	assert.Equal(t, "eno1", n3.BootInterface().Name)
	assert.Equal(t, "d0:8e:79:00:00:03", n3.InterfaceBMC().MAC.String(), "the BMC is kept")

	n1 := hosts[1]
	assert.Equal(t, "b8:ca:3a:00:00:01", n1.BootInterface().MAC.String(), "the boot port stays first")
	assert.Len(t, n1.Interfaces, 4, "the new ports are added after it")
	assert.Equal(t, "b8:ca:3a:00:00:01", importHosts()[0].Interfaces[0].MAC.String(), "the stored node is not changed in place")

	assert.Equal(t, "cpn-a01-02.cluster", hosts[2].BootInterface().FQDN)
}

func TestPlanImportBootRowOnBMC(t *testing.T) {
	p := planImport(importHosts(), []NodeImportRow{{Node: "cpn-a01-01", MAC: "d0:8e:79:00:00:01"}}, bmcOpts())
	assert.Equal(t, ImportConflict, p.rows[0].Action)
	assert.Equal(t, "MAC is already on another port of cpn-a01-01", p.rows[0].Reason)
}

func TestPlanImportRanges(t *testing.T) {
	t.Run("after_highest starts past the highest held", func(t *testing.T) {
		o := bmcOpts()
		o.afterHighest = true
		p := planImport(importHosts(), []NodeImportRow{{Node: "n1", MAC: "aa:aa:aa:00:00:01", BMC: true}}, o)
		assert.Equal(t, "10.1.0.4/22", p.rows[0].IP)
	})

	t.Run("a full range is a conflict on the rows that do not fit", func(t *testing.T) {
		o := importOpts{bmc: netip.MustParsePrefix("10.9.0.0/30")}
		rows := []NodeImportRow{
			{Node: "n1", MAC: "aa:aa:aa:00:00:01", BMC: true},
			{Node: "n2", MAC: "aa:aa:aa:00:00:02", BMC: true},
			{Node: "n3", MAC: "aa:aa:aa:00:00:03", BMC: true},
		}
		p := planImport(nil, rows, o)
		assert.Equal(t, "10.9.0.1/30", p.rows[0].IP)
		assert.Equal(t, "10.9.0.2/30", p.rows[1].IP)
		assert.Equal(t, ImportConflict, p.rows[2].Action)
		assert.Equal(t, "no free address left in 10.9.0.0/30", p.rows[2].Reason)
	})

	t.Run("no range for the kind of port", func(t *testing.T) {
		p := planImport(nil, []NodeImportRow{{Node: "n1", MAC: "aa:aa:aa:00:00:01"}}, importOpts{bmc: netip.MustParsePrefix("10.1.0.0/24")})
		assert.Equal(t, "no address given and no boot range to take one from", p.rows[0].Reason)
	})

	t.Run("given addresses are never handed out again", func(t *testing.T) {
		rows := []NodeImportRow{
			{Node: "n1", MAC: "aa:aa:aa:00:00:01", BMC: true},
			{Node: "n2", MAC: "aa:aa:aa:00:00:02", BMC: true, IP: "10.1.0.2/22"},
		}
		p := planImport(importHosts(), rows, bmcOpts())
		assert.Equal(t, "10.1.0.4/22", p.rows[0].IP)
	})

	t.Run("an unchanged port with no address is given one", func(t *testing.T) {
		hosts := model.HostList{{Name: "n1", Interfaces: []*model.NetInterface{port("aa:aa:aa:00:00:01", "", true)}}}
		p := planImport(hosts, []NodeImportRow{{Node: "n1", MAC: "aa:aa:aa:00:00:01", BMC: true}}, bmcOpts())
		assert.Equal(t, ImportReplace, p.rows[0].Action)
		assert.Equal(t, "10.1.0.1/22", p.rows[0].IP)
	})
}

func TestPlanWithPingSkipsAnswers(t *testing.T) {
	asked := fakePing(t, "10.1.0.2", "10.1.0.4")
	rows := []NodeImportRow{
		{Node: "cpn-a01-02", MAC: "d0:8e:79:00:00:02", BMC: true},
		{Node: "cpn-a01-04", MAC: "d0:8e:79:00:00:04", BMC: true},
	}
	p, pinged, perr := planWithPing(context.Background(), importHosts(), rows, bmcOpts())
	require.True(t, pinged, perr)
	assert.Equal(t, "10.1.0.5/22", p.rows[0].IP)
	assert.Equal(t, "10.1.0.6/22", p.rows[1].IP)
	assert.NotContains(t, *asked, netip.MustParseAddr("10.1.0.1"), "a port already holding its address is not pinged")
}

func TestPlanWithPingReportsFailure(t *testing.T) {
	old := pingAddrs
	pingAddrs = func(context.Context, []netip.Addr, int, time.Duration) (map[netip.Addr]bool, error) {
		return nil, errors.New("operation not permitted")
	}
	t.Cleanup(func() { pingAddrs = old })

	p, pinged, perr := planWithPing(context.Background(), nil, []NodeImportRow{{Node: "n1", MAC: "aa:aa:aa:00:00:01", BMC: true}}, bmcOpts())
	assert.False(t, pinged)
	assert.Equal(t, "operation not permitted", perr)
	assert.Equal(t, "10.1.0.1/22", p.rows[0].IP, "addresses are still planned")
}

func runImport(t *testing.T, h *Handler, req NodeImportRequest) (*NodeImportResponse, error) {
	t.Helper()
	return h.NodeImport(fuego.NewMockContext(req))
}

func TestNodeImportCommit(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)
	require.NoError(t, h.DB.StoreHosts(importHosts()))

	rows := []NodeImportRow{
		{Node: "cpn-a01-02", MAC: "d0:8e:79:00:00:02", BMC: true, FQDN: "bmc-cpn-a01-02.cluster", Tags: []string{"a01"}},
		{Node: "cpn-a01-05", MAC: "d0:8e:79:00:00:05", BMC: true, Tags: []string{"a01"}},
	}
	preview, err := runImport(t, h, NodeImportRequest{DryRun: true, Ranges: NodeImportRanges{BMC: "10.1.0.0/22"}, Rows: rows})
	require.NoError(t, err)
	assert.False(t, preview.Committed)
	assert.Equal(t, NodeImportSummary{Add: 1, Create: 1}, preview.Summary)

	// the client commits the previewed addresses, so what is stored is what was shown
	for i := range rows {
		rows[i].IP = preview.Rows[i].IP
	}
	res, err := runImport(t, h, NodeImportRequest{Rows: rows})
	require.NoError(t, err)
	assert.True(t, res.Committed)

	n2, err := h.DB.LoadHostFromName("cpn-a01-02")
	require.NoError(t, err)
	require.NotNil(t, n2.InterfaceBMC())
	assert.Equal(t, "10.1.0.2/22", n2.InterfaceBMC().IP.String())
	assert.Equal(t, "bmc-cpn-a01-02.cluster", n2.InterfaceBMC().FQDN)
	assert.Equal(t, "b8:ca:3a:00:00:02", n2.BootInterface().MAC.String(), "the boot port is kept")
	assert.Contains(t, n2.Tags, "a01")

	n5, err := h.DB.LoadHostFromName("cpn-a01-05")
	require.NoError(t, err)
	assert.Equal(t, "d0:8e:79:00:00:05", n5.InterfaceBMC().MAC.String())

	again, err := runImport(t, h, NodeImportRequest{DryRun: true, Ranges: NodeImportRanges{BMC: "10.1.0.0/22"}, Rows: rows})
	require.NoError(t, err)
	assert.Equal(t, NodeImportSummary{Unchanged: 2}, again.Summary, "running the same import again changes nothing")
}

func TestNodeImportCommitRefusesConflicts(t *testing.T) {
	h := newTestHandler(t)
	require.NoError(t, h.DB.StoreHosts(importHosts()))

	_, err := runImport(t, h, NodeImportRequest{Ranges: NodeImportRanges{BMC: "10.1.0.0/22"}, Rows: []NodeImportRow{
		{Node: "cpn-a01-08", MAC: "aa:aa:aa:00:00:08", BMC: true},
		{Node: "cpn-a01-09", MAC: "d0:8e:79:00:00:01", BMC: true},
	}})
	var herr fuego.HTTPError
	require.ErrorAs(t, err, &herr)
	assert.Equal(t, 409, herr.Status)

	_, err = h.DB.LoadHostFromName("cpn-a01-08")
	assert.Error(t, err, "nothing is stored when any row conflicts")
}

func TestNodeImportPingedDryRunBadRows(t *testing.T) {
	h := newTestHandler(t)
	fakePing(t)

	res, err := runImport(t, h, NodeImportRequest{DryRun: true, Ping: true, Rows: []NodeImportRow{
		{Node: "n1", MAC: "zz", IP: "10.9.0.7"},
		{Node: "n2", MAC: "aa:aa:aa:00:00:02", IP: "10.9.0.300"},
	}})
	require.NoError(t, err)
	assert.Equal(t, NodeImportSummary{Conflict: 2}, res.Summary)
	assert.Contains(t, res.Rows[0].Reason, "is not a MAC address")
	assert.Equal(t, "10.9.0.7", res.Rows[0].IP, "a row that conflicts keeps what was given")
	assert.Contains(t, res.Rows[1].Reason, "is not an IPv4 address")
}

func TestPlanImportOverlappingRanges(t *testing.T) {
	both := netip.MustParsePrefix("10.9.0.0/24")
	rows := []NodeImportRow{
		{Node: "n1", MAC: "aa:aa:aa:00:00:01", BMC: true},
		{Node: "n1", MAC: "aa:aa:aa:00:00:02"},
		{Node: "n2", MAC: "aa:aa:aa:00:00:03", BMC: true},
	}
	p := planImport(nil, rows, importOpts{bmc: both, boot: both})
	assert.Equal(t, []string{"10.9.0.1/24", "10.9.0.2/24", "10.9.0.3/24"}, []string{p.rows[0].IP, p.rows[1].IP, p.rows[2].IP}, "ranges that overlap share what has been handed out")
	assert.Zero(t, p.summary().Conflict)
}

func TestPlanWithPingOverlappingRanges(t *testing.T) {
	asked := fakePing(t, "10.9.0.2")
	both := netip.MustParsePrefix("10.9.0.0/24")
	rows := []NodeImportRow{
		{Node: "n1", MAC: "aa:aa:aa:00:00:01", BMC: true},
		{Node: "n1", MAC: "aa:aa:aa:00:00:02"},
	}
	p, pinged, perr := planWithPing(context.Background(), nil, rows, importOpts{bmc: both, boot: both})
	require.True(t, pinged, perr)
	assert.Equal(t, "10.9.0.1/24", p.rows[0].IP)
	assert.Equal(t, "10.9.0.3/24", p.rows[1].IP, "past the address the BMC row took and the one that answered")
	assert.Contains(t, *asked, netip.MustParseAddr("10.9.0.3"))
}

func TestPlanImportAfterHighestAtTopOfRange(t *testing.T) {
	hosts := model.HostList{{Name: "n0", Interfaces: []*model.NetInterface{port("aa:aa:aa:00:00:00", "10.9.0.254/24", true)}}}
	o := importOpts{bmc: netip.MustParsePrefix("10.9.0.0/24"), afterHighest: true}
	p := planImport(hosts, []NodeImportRow{{Node: "n1", MAC: "aa:aa:aa:00:00:01", BMC: true}}, o)
	assert.Equal(t, ImportConflict, p.rows[0].Action)
	assert.Equal(t, "no free address in 10.9.0.0/24 after 10.9.0.254, the highest in use", p.rows[0].Reason, "the range is not full, so it isn't called full")
}

func TestPlanImportSkipsTheGateway(t *testing.T) {
	withSubnets(t, "10.9.0.1/24")
	p := planImport(nil, []NodeImportRow{{Node: "n1", MAC: "aa:aa:aa:00:00:01", BMC: true}}, importOpts{bmc: netip.MustParsePrefix("10.9.0.0/24")})
	assert.Equal(t, "10.9.0.2/24", p.rows[0].IP)
}

func TestPlanWithPingNeverAssignsUnpinged(t *testing.T) {
	asked := map[netip.Addr]bool{}
	old := pingAddrs
	pingAddrs = func(_ context.Context, addrs []netip.Addr, _ int, _ time.Duration) (map[netip.Addr]bool, error) {
		out := map[netip.Addr]bool{}
		for _, a := range addrs {
			asked[a], out[a] = true, true
		}
		return out, nil
	}
	t.Cleanup(func() { pingAddrs = old })

	rows := []NodeImportRow{{Node: "n1", MAC: "aa:aa:aa:00:00:01", BMC: true}}
	p, pinged, perr := planWithPing(context.Background(), nil, rows, importOpts{bmc: netip.MustParsePrefix("10.9.0.0/24")})
	require.True(t, pinged, perr)
	assert.NotEmpty(t, asked)
	assert.Equal(t, ImportConflict, p.rows[0].Action, "an address that was not pinged is not handed out as pinged")
	assert.False(t, p.rows[0].Assigned)
	assert.Contains(t, p.rows[0].Reason, "answered a ping")
}

type fakeSwitch struct {
	macs      model.MACTable
	status    map[string]string
	macErr    error
	statusErr error
}

func (f fakeSwitch) GetMACTable(context.Context) (model.MACTable, error) { return f.macs, f.macErr }
func (fakeSwitch) Close() error                                          { return nil }
func (f fakeSwitch) GetInterfaceDescriptions(context.Context) (map[string]string, error) {
	return f.status, f.statusErr
}
func (fakeSwitch) GetLLDPNeighbors(context.Context) (model.LLDPNeighbors, error) { return nil, nil }

func mac(s string) net.HardwareAddr { hw, _ := net.ParseMAC(s); return hw }

func TestScanSwitch(t *testing.T) {
	fs := fakeSwitch{
		macs: model.MACTable{
			{Ifname: "Ethernet1", VLAN: "3001", MAC: mac("d0:8e:79:00:00:01")},
			{Ifname: "Ethernet4", VLAN: "3001", MAC: mac("d0:8e:79:00:00:04")},
			{Ifname: "Ethernet4", VLAN: "3001", MAC: mac("b8:ca:3a:00:00:04")},
			{Ifname: "Port-Channel1", VLAN: "1", MAC: mac("aa:aa:aa:00:00:01")},
		},
		status: map[string]string{"Ethernet1": "", "Ethernet2": "spare", "Ethernet10": "", "Vlan3001": ""},
	}
	for i := range 10 {
		fs.macs = append(fs.macs, &model.MACTableEntry{Ifname: "Ethernet49/1", VLAN: "1", MAC: net.HardwareAddr{0xaa, 0, 0, 0, 1, byte(i)}})
	}

	old := newSwitch
	newSwitch = func(*model.Host, tors.Config) (tors.NetworkSwitch, error) { return fs, nil }
	t.Cleanup(func() { newSwitch = old })

	owners := map[string]macOwner{"d0:8e:79:00:00:01": {node: "cpn-a01-01", bmc: true}}
	res := scanSwitch(context.Background(), &model.Host{Name: "swe-a01", Tags: []string{"arista"}}, tors.Config{}, owners)

	require.Empty(t, res.Error)
	names := []string{}
	for _, p := range res.Ports {
		names = append(names, p.Ifname)
	}
	assert.Equal(t, []string{"Ethernet1", "Ethernet2", "Ethernet4", "Ethernet10", "Ethernet49/1"}, names, "physical ports only, in natural order")

	assert.Equal(t, "cpn-a01-01", res.Ports[0].MACs[0].Owner)
	assert.True(t, res.Ports[0].MACs[0].BMC)
	assert.Equal(t, "spare", res.Ports[1].Description)
	assert.Empty(t, res.Ports[1].MACs, "a port with no MACs is still listed")
	assert.Len(t, res.Ports[2].MACs, 2, "both shared-LOM MACs are reported")
	assert.Len(t, res.Ports[4].MACs, 10)
}

func TestScanSwitchErrors(t *testing.T) {
	old := newSwitch
	t.Cleanup(func() { newSwitch = old })

	newSwitch = func(*model.Host, tors.Config) (tors.NetworkSwitch, error) { return nil, tors.ErrUnknownNOS }
	res := scanSwitch(context.Background(), &model.Host{Name: "swe-a01"}, tors.Config{}, nil)
	assert.Equal(t, "failed to determine switch NOS", res.Error)
	assert.Equal(t, tors.CodeUnknownNOS, res.ErrorCode)

	newSwitch = func(*model.Host, tors.Config) (tors.NetworkSwitch, error) {
		return fakeSwitch{macs: model.MACTable{}, statusErr: errors.New("timeout")}, nil
	}
	res = scanSwitch(context.Background(), &model.Host{Name: "swe-a01"}, tors.Config{}, nil)
	assert.Empty(t, res.Error, "an interface status failure leaves the scan usable")
	assert.Equal(t, []string{"interface status: timeout"}, res.Warnings)

	newSwitch = func(*model.Host, tors.Config) (tors.NetworkSwitch, error) {
		return fakeSwitch{macErr: errors.New("connection refused")}, nil
	}
	res = scanSwitch(context.Background(), &model.Host{Name: "swe-a01"}, tors.Config{}, nil)
	assert.Equal(t, "failed to read the MAC table: connection refused", res.Error)
	assert.Equal(t, tors.CodeUnreachable, res.ErrorCode)
}
