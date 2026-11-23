// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/go-faster/jx"
	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/client"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

type fakeBmcContext struct {
	query    map[string]string
	ctx      context.Context
	recorder *httptest.ResponseRecorder
}

func newFakeBmcContext(query map[string]string) *fakeBmcContext {
	return &fakeBmcContext{
		query:    query,
		ctx:      context.WithValue(context.Background(), ContextKeyUsername, "tester"),
		recorder: httptest.NewRecorder(),
	}
}

func (f *fakeBmcContext) QueryParam(name string) string { return f.query[name] }

// an absent parameter reads as false, the default the routes declare
func (f *fakeBmcContext) QueryParamBoolErr(name string) (bool, error) {
	if f.query[name] == "" {
		return false, nil
	}
	return strconv.ParseBool(f.query[name])
}
func (f *fakeBmcContext) Context() context.Context      { return f.ctx }
func (f *fakeBmcContext) Response() http.ResponseWriter { return f.recorder }

// TestBmcSelect pins which nodes a bmc request reaches. These routes can power off machines, so a selector that widens by accident is the worst failure they have
func TestBmcSelect(t *testing.T) {
	h := newTestHandler(t)

	tagged := tests.HostFactory.MustCreate().(*model.Host)
	untagged := tests.HostFactory.MustCreate().(*model.Host)
	if err := h.DB.StoreHosts(model.HostList{tagged, untagged}); err != nil {
		t.Fatal(err)
	}

	ns, err := nodeset.NewNodeSet(tagged.Name)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.DB.TagHosts(ns, []string{"gpu"}); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name    string
		all     string
		nodeset string
		tags    string
		want    []string
		wantErr bool
	}{
		{name: "no selector is rejected", wantErr: true},
		{name: "all reaches every node", all: "true", want: []string{tagged.Name, untagged.Name}},
		{name: "all is narrowed by tags", all: "true", tags: "gpu", want: []string{tagged.Name}},
		{name: "all is narrowed by a nodeset", all: "true", nodeset: untagged.Name, want: []string{untagged.Name}},
		{name: "all with an unknown tag reaches nothing", all: "true", tags: "typo", wantErr: true},
		{name: "tags alone", tags: "gpu", want: []string{tagged.Name}},
		{name: "unknown tag alone reaches nothing", tags: "typo", wantErr: true},
		{name: "nodeset with an unknown tag reaches nothing", nodeset: tagged.Name + "," + untagged.Name, tags: "typo", wantErr: true},
		{name: "nodeset and tags intersect", nodeset: tagged.Name + "," + untagged.Name, tags: "gpu", want: []string{tagged.Name}},
		{name: "nodeset and tags with no overlap reach nothing", nodeset: untagged.Name, tags: "gpu"},
		{name: "a name not in the database fails the request", nodeset: tagged.Name + ",notindb01", wantErr: true},
		{name: "a name not in the database fails even with all", all: "true", nodeset: tagged.Name + ",notindb01", wantErr: true},
		{name: "a name not in the database fails even when the tag would exclude it", nodeset: tagged.Name + ",notindb01", tags: "gpu", wantErr: true},
		{name: "an unparseable nodeset fails the request", nodeset: "cpn-[", wantErr: true},
		{name: "all that is not a boolean fails the request", all: "yes", nodeset: tagged.Name, wantErr: true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			hosts, err := h.bmcSelect(newFakeBmcContext(map[string]string{"all": tt.all, "nodeset": tt.nodeset, "tags": tt.tags}))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}

			got := make([]string, 0, len(hosts))
			for _, host := range hosts {
				got = append(got, host.Name)
			}
			slices.Sort(got)

			want := slices.Clone(tt.want)
			slices.Sort(want)

			if !slices.Equal(got, want) {
				t.Errorf("hosts = %v, want %v", got, want)
			}
		})
	}
}

// TestBmcNodes covers the two ways a request can resolve to nothing, which have to be errors because an empty success would exit 0 having done nothing
func TestBmcNodes(t *testing.T) {
	h := newTestHandler(t)

	withBMC := tests.HostFactory.MustCreate().(*model.Host)
	noBMC := tests.HostFactory.MustCreate().(*model.Host)
	for _, nic := range noBMC.Interfaces {
		nic.BMC = false
	}
	if err := h.DB.StoreHosts(model.HostList{withBMC, noBMC}); err != nil {
		t.Fatal(err)
	}

	assertNotFound := func(t *testing.T, err error) {
		t.Helper()
		var httpErr fuego.HTTPError
		if !errors.As(err, &httpErr) || httpErr.Status != http.StatusNotFound {
			t.Fatalf("err = %v, want a 404", err)
		}
	}

	t.Run("a named node without a bmc interface is reported, not contacted", func(t *testing.T) {
		nodes, err := h.bmcNodes(newFakeBmcContext(map[string]string{"nodeset": withBMC.Name + "," + noBMC.Name}))
		if err != nil {
			t.Fatal(err)
		}
		if len(nodes.hosts) != 1 || nodes.hosts[0].Name != withBMC.Name {
			t.Errorf("hosts = %+v, want only %s", nodes.hosts, withBMC.Name)
		}
		if !slices.Equal(nodes.noBMC, []string{noBMC.Name}) {
			t.Errorf("noBMC = %v, want %s", nodes.noBMC, noBMC.Name)
		}
	})

	t.Run("a node without a bmc interface that all swept in is left out", func(t *testing.T) {
		nodes, err := h.bmcNodes(newFakeBmcContext(map[string]string{"all": "true"}))
		if err != nil {
			t.Fatal(err)
		}
		if len(nodes.hosts) != 1 || len(nodes.noBMC) != 0 {
			t.Errorf("hosts = %+v, noBMC = %v, want only %s", nodes.hosts, nodes.noBMC, withBMC.Name)
		}
	})

	t.Run("hosts are in natural order, which is the order templates number them in", func(t *testing.T) {
		h := newTestHandler(t)
		for _, name := range []string{"cpn-f01-10", "cpn-f01-9", "cpn-f01-2"} {
			host := tests.HostFactory.MustCreate().(*model.Host)
			host.Name = name
			if err := h.DB.StoreHost(host); err != nil {
				t.Fatal(err)
			}
		}

		nodes, err := h.bmcNodes(newFakeBmcContext(map[string]string{"all": "true"}))
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, host := range nodes.hosts {
			names = append(names, host.Name)
		}
		if want := []string{"cpn-f01-2", "cpn-f01-9", "cpn-f01-10"}; !slices.Equal(names, want) {
			t.Errorf("hosts = %v, want %v", names, want)
		}
	})

	t.Run("only nodes without a bmc interface", func(t *testing.T) {
		_, err := h.bmcNodes(newFakeBmcContext(map[string]string{"nodeset": noBMC.Name}))
		assertNotFound(t, err)
	})

	t.Run("only names not in the database", func(t *testing.T) {
		_, err := h.bmcNodes(newFakeBmcContext(map[string]string{"nodeset": "notindb01"}))
		assertNotFound(t, err)
	})
}

// a named node that cannot be contacted still gets a row, so a run over a nodeset accounts for every name in it
func TestRunBmcReportsNamedNodeWithoutBMC(t *testing.T) {
	h := newTestHandler(t)

	withBMC := tests.HostFactory.MustCreate().(*model.Host)
	withBMC.Name = "cpn-1"
	noBMC := tests.HostFactory.MustCreate().(*model.Host)
	noBMC.Name = "cpn-2"
	for _, nic := range noBMC.Interfaces {
		nic.BMC = false
	}
	if err := h.DB.StoreHosts(model.HostList{withBMC, noBMC}); err != nil {
		t.Fatal(err)
	}

	var contacted []string
	res, err := runAck(h, newFakeBmcContext(map[string]string{"nodeset": "cpn-[1-2]"}), "bmc.power", "bmc power On",
		func(_ context.Context, targets []redfish.Target) []redfish.NodeResult[redfish.Ack] {
			results := make([]redfish.NodeResult[redfish.Ack], 0, len(targets))
			for _, target := range targets {
				contacted = append(contacted, target.Name)
				results = append(results, redfish.NodeResult[redfish.Ack]{Node: target.Name, Success: true})
			}
			return results
		})
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(contacted, []string{"cpn-1"}) {
		t.Errorf("contacted %v, want only cpn-1", contacted)
	}
	if res.Summary.Total != 2 || res.Summary.Failed != 1 || res.Summary.FailedNodes != "cpn-2" {
		t.Errorf("summary = %+v, want cpn-2 counted as failed", res.Summary)
	}
	if row := res.Results[1]; row.Node != "cpn-2" || row.Success || !strings.Contains(row.Message, "no interface marked bmc") {
		t.Errorf("row = %+v, want cpn-2 failed for having no bmc interface", row)
	}

	event, err := h.DB.LoadEvent(listEvents(t, h)[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(event.Nodes) != 2 {
		t.Errorf("event has %d nodes, want both", len(event.Nodes))
	}
}

// TestRunBmcMistypedNodeContactsNothing is the behaviour that matters most for a command that powers machines off: one bad name stops the whole run
func TestRunBmcMistypedNodeContactsNothing(t *testing.T) {
	h := newTestHandler(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	if err := h.DB.StoreHost(host); err != nil {
		t.Fatal(err)
	}

	contacted := false
	_, err := runAck(h, newFakeBmcContext(map[string]string{"nodeset": host.Name + ",notindb01,notindb02"}), "bmc.power", "bmc power ForceOff",
		func(context.Context, []redfish.Target) []redfish.NodeResult[redfish.Ack] {
			contacted = true
			return nil
		})

	var httpErr fuego.HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != http.StatusNotFound {
		t.Fatalf("err = %v, want a 404", err)
	}
	if !strings.Contains(httpErr.Detail, "notindb[01-02]") {
		t.Errorf("detail = %q, it has to name the bad nodes", httpErr.Detail)
	}
	if contacted {
		t.Error("a bmc was contacted despite the bad name")
	}
	if events := listEvents(t, h); len(events) != 0 {
		t.Errorf("wrote %d events for a run that did nothing", len(events))
	}
}

func TestRunBmc(t *testing.T) {
	h := newTestHandler(t)

	for _, name := range []string{"cpn-2", "cpn-3", "cpn-10"} {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = name
		if err := h.DB.StoreHost(host); err != nil {
			t.Fatal(err)
		}
	}

	// like redfish.Run, it reports on every target, though out of order
	stub := func(_ context.Context, targets []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		outcome := map[string]redfish.NodeResult[redfish.Ack]{
			"cpn-10": {Node: "cpn-10", Success: true, Data: &redfish.Ack{}},
			"cpn-3":  {Node: "cpn-3", Message: "connection refused"},
			"cpn-2":  {Node: "cpn-2", Message: "timed out"},
		}
		results := make([]redfish.NodeResult[redfish.Ack], 0, len(targets))
		for i := len(targets) - 1; i >= 0; i-- {
			results = append(results, outcome[targets[i].Name])
		}
		return results
	}

	c := newFakeBmcContext(map[string]string{"nodeset": "cpn-[2-3],cpn-10"})

	res, err := runAck(h, c, "bmc.power", "bmc power ForceOff", stub)
	if err != nil {
		t.Fatal(err)
	}

	gotNodes := []string{res.Results[0].Node, res.Results[1].Node, res.Results[2].Node}
	if want := []string{"cpn-2", "cpn-3", "cpn-10"}; !slices.Equal(gotNodes, want) {
		t.Errorf("rows are in order %v, want natural order %v", gotNodes, want)
	}

	if res.Summary.Total != 3 || res.Summary.Succeeded != 1 || res.Summary.Failed != 2 {
		t.Errorf("summary = %+v", res.Summary)
	}
	if res.Summary.FailedNodes != "cpn-[2-3]" {
		t.Errorf("failed nodes = %q, want a nodeset that can be pasted into another command", res.Summary.FailedNodes)
	}
	if got := c.recorder.Header().Get("X-Grendel-Failed"); got != "2" {
		t.Errorf("X-Grendel-Failed = %q", got)
	}

	events := listEvents(t, h)
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	event, err := h.DB.LoadEvent(events[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != model.EventKindBmc || event.Action != "bmc.power" || event.Summary != "bmc power ForceOff" {
		t.Errorf("event is %s %q %q", event.Kind, event.Action, event.Summary)
	}
	if event.State != model.EventStateDone || event.EndedAt == nil {
		t.Errorf("state = %q, ended = %v, a finished run is closed", event.State, event.EndedAt)
	}
	if event.Severity != model.SeverityWarning {
		t.Errorf("severity = %q for a partly failed run", event.Severity)
	}
	if event.Detail["run"] != res.RunID {
		t.Errorf("run = %q, want the response's run id %q", event.Detail["run"], res.RunID)
	}
	if want := (model.EventCounts{Total: 3, Succeeded: 1, Failed: 2}); events[0].Counts != want {
		t.Errorf("counts = %+v, want %+v", events[0].Counts, want)
	}
	// each node's row says how it went and, when it failed, why
	want := map[string]model.EventNode{
		"cpn-2":  {State: model.EventNodeFailed, Message: "timed out"},
		"cpn-3":  {State: model.EventNodeFailed, Message: "connection refused"},
		"cpn-10": {State: model.EventNodeSucceeded},
	}
	if len(event.Nodes) != len(want) {
		t.Fatalf("got %d nodes, want %d", len(event.Nodes), len(want))
	}
	for _, n := range event.Nodes {
		if n.State != want[n.Node].State || n.Message != want[n.Node].Message {
			t.Errorf("%s is %q %q, want %q %q", n.Node, n.State, n.Message, want[n.Node].State, want[n.Node].Message)
		}
	}
}

// TestRunBmcReadsStayOutOfTheEventLog matters because the log keeps few events, and status polling would push the power actions out of it
func TestRunBmcReadsStayOutOfTheEventLog(t *testing.T) {
	h := newTestHandler(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	if err := h.DB.StoreHost(host); err != nil {
		t.Fatal(err)
	}

	_, err := runBmc(h, newFakeBmcContext(map[string]string{"nodeset": host.Name}), "bmc status", nil,
		func(context.Context, []redfish.Target) []redfish.NodeResult[model.RedfishSystem] { return nil },
		func(d *model.RedfishSystem) *model.RedfishSystem { return d },
	)
	if err != nil {
		t.Fatal(err)
	}

	if events := listEvents(t, h); len(events) != 0 {
		t.Errorf("a read wrote %d events", len(events))
	}
}

// TestGeneratedClientDecodesResponses sends what the handlers actually marshal through the generated client, because a spec that disagrees with the wire only shows up as the CLI failing to decode
func TestGeneratedClientDecodesResponses(t *testing.T) {
	summary := model.BmcSummary{Total: 2, Succeeded: 1, Failed: 1, FailedNodes: "cpn-2"}

	t.Run("status with a non dell node and a failed one", func(t *testing.T) {
		body, err := json.Marshal(model.BmcResponse[model.RedfishSystem]{RunID: "run", Summary: summary, Results: []model.BmcResult[model.RedfishSystem]{
			{Node: "cpn-1", Success: true, Data: &model.RedfishSystem{Name: "cpn-1", SerialNumber: "S1"}},
			{Node: "cpn-2", Message: "boom"},
		}})
		if err != nil {
			t.Fatal(err)
		}

		var res client.BmcResponseModelRedfishSystem
		if err := res.Decode(jx.DecodeBytes(body)); err != nil {
			t.Fatalf("decoding %s: %v", body, err)
		}

		if got := res.Summary.FailedNodes; got != "cpn-2" {
			t.Errorf("failed nodes = %q", got)
		}

		reached := res.Results[0].Data.Value
		if reached.SerialNumber != "S1" {
			t.Errorf("serial = %q", reached.SerialNumber)
		}
		if !reached.OemDell.Null {
			t.Errorf("a non dell node should decode a null oem block, got %+v", reached.OemDell)
		}
		if !res.Results[0].Success {
			t.Error("the reached node should decode as a success")
		}
		if res.Results[1].Success {
			t.Error("the failed node should decode as a failure")
		}
		if !res.Results[1].Data.Null {
			t.Error("a failed node should decode null data")
		}
	})

	t.Run("action", func(t *testing.T) {
		body, err := json.Marshal(model.BmcResponse[model.BmcAck]{RunID: "run", Summary: summary, Results: []model.BmcResult[model.BmcAck]{
			{Node: "cpn-1", Success: true, Message: "cleared 1 job(s)"},
			{Node: "cpn-2", Message: "cleared 0 of 1 job(s), failed on JID_1: boom"},
		}})
		if err != nil {
			t.Fatal(err)
		}

		var res client.BmcResponseModelBmcAck
		if err := res.Decode(jx.DecodeBytes(body)); err != nil {
			t.Fatalf("decoding %s: %v", body, err)
		}

		if got := res.Results[0].Message; got != "cleared 1 job(s)" {
			t.Errorf("message = %q", got)
		}
	})
}

func TestValidJobID(t *testing.T) {
	for _, id := range []string{"JID_123456789", "RID_1", "job-1.2:3"} {
		if !validJobID(id) {
			t.Errorf("%q should be valid", id)
		}
	}

	for _, id := range []string{"", ".", "..", "../../AccountService/Accounts/3", "a/b", "a b", "a%2Fb", "a?b"} {
		if validJobID(id) {
			t.Errorf("%q should be rejected", id)
		}
	}
}

func listEvents(t *testing.T, h *Handler) model.EventList {
	t.Helper()
	events, err := h.DB.Events(model.EventFilter{})
	if err != nil {
		t.Fatal(err)
	}
	return events
}

// a bmc target is always a graceful restart, so a body that also asks for a reset type must not be answered with one
func TestPowerRequestBmcTargetTakesNothingElse(t *testing.T) {
	req, err := powerRequest(BmcPowerBody{Target: "bmc"})
	if err != nil || req.Target != redfish.TargetBMC {
		t.Fatalf("req = %+v, err = %v, want a bmc restart", req, err)
	}

	for _, body := range []BmcPowerBody{
		{Target: "bmc", Action: "ForceOff"},
		{Target: "bmc", BootOverride: "Pxe"},
	} {
		var httpErr fuego.HTTPError
		if _, err := powerRequest(body); !errors.As(err, &httpErr) || httpErr.Status != http.StatusBadRequest {
			t.Errorf("%+v: err = %v, want a 400", body, err)
		}
	}
}
