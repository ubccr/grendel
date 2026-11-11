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
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			c := fuego.NewMockContextNoBody()
			c.SetQueryParam("all", tt.all)
			c.SetQueryParam("nodeset", tt.nodeset)
			c.SetQueryParam("tags", tt.tags)

			hosts, err := h.bmcSelect(c)
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

// TestBmcTargets covers the two ways a request can resolve to nothing, which have to be errors because an empty success would exit 0 having done nothing
func TestBmcTargets(t *testing.T) {
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

	t.Run("a node without a bmc interface is ignored", func(t *testing.T) {
		targets, err := h.bmcTargets(newFakeBmcContext(map[string]string{"nodeset": withBMC.Name + "," + noBMC.Name}))
		if err != nil {
			t.Fatal(err)
		}
		if len(targets) != 1 || targets[0].Name != withBMC.Name {
			t.Errorf("targets = %+v, want only %s", targets, withBMC.Name)
		}
	})

	t.Run("only nodes without a bmc interface", func(t *testing.T) {
		_, err := h.bmcTargets(newFakeBmcContext(map[string]string{"nodeset": noBMC.Name}))
		assertNotFound(t, err)
	})

	t.Run("only names not in the database", func(t *testing.T) {
		_, err := h.bmcTargets(newFakeBmcContext(map[string]string{"nodeset": "notindb01"}))
		assertNotFound(t, err)
	})
}

// TestRunBmcMistypedNodeContactsNothing is the behaviour that matters most for a command that powers machines off: one bad name stops the whole run
func TestRunBmcMistypedNodeContactsNothing(t *testing.T) {
	h := newTestHandler(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	if err := h.DB.StoreHost(host); err != nil {
		t.Fatal(err)
	}

	contacted := false
	_, err := runAck(h, newFakeBmcContext(map[string]string{"nodeset": host.Name + ",notindb01,notindb02"}), "bmc power ForceOff",
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
	if events := h.Events.GetEvents(); len(events) != 0 {
		t.Errorf("wrote %d events for a run that did nothing", len(events))
	}
}

func TestRunBmc(t *testing.T) {
	h := newTestHandler(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	if err := h.DB.StoreHost(host); err != nil {
		t.Fatal(err)
	}

	stub := func(context.Context, []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return []redfish.NodeResult[redfish.Ack]{
			{Node: "cpn-10", Success: true, Data: &redfish.Ack{}},
			{Node: "cpn-3", Message: "connection refused"},
			{Node: "cpn-2", Message: "timed out"},
		}
	}

	c := newFakeBmcContext(map[string]string{"nodeset": host.Name})

	res, err := runAck(h, c, "bmc power ForceOff", stub)
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

	events := h.Events.GetEvents()
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}

	event := events[0]
	if event.Severity != model.SeverityWarning.String() {
		t.Errorf("severity = %q for a partly failed run", event.Severity)
	}
	if want := "bmc power ForceOff: 1 succeeded, 2 failed of 3 (run " + res.RunID + ")"; event.Message != want {
		t.Errorf("message = %q, want %q", event.Message, want)
	}
	if len(event.JobMessages) != 3 {
		t.Fatalf("got %d job messages, want one per node", len(event.JobMessages))
	}
	for _, m := range event.JobMessages {
		want := "error"
		if m.Host == "cpn-10" {
			want = "success"
		}
		if m.Status != want {
			t.Errorf("%s status = %q, want %q", m.Host, m.Status, want)
		}
		if m.Msg != "" {
			t.Errorf("%s carries error detail %q, the event only records the outcome", m.Host, m.Msg)
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

	_, err := runBmc(h, newFakeBmcContext(map[string]string{"nodeset": host.Name}), "bmc status", false,
		func(context.Context, []redfish.Target) []redfish.NodeResult[model.RedfishSystem] { return nil },
		func(d *model.RedfishSystem) *model.RedfishSystem { return d },
	)
	if err != nil {
		t.Fatal(err)
	}

	if events := h.Events.GetEvents(); len(events) != 0 {
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

		if got := res.Summary.Value.FailedNodes.Value; got != "cpn-2" {
			t.Errorf("failed nodes = %q", got)
		}

		reached := res.Results.Value[0].Value.Data.Value
		if reached.SerialNumber.Value != "S1" {
			t.Errorf("serial = %q", reached.SerialNumber.Value)
		}
		if !reached.OemDell.Null {
			t.Errorf("a non dell node should decode a null oem block, got %+v", reached.OemDell)
		}
		if !res.Results.Value[0].Value.Success.Value {
			t.Error("the reached node should decode as a success")
		}
		if res.Results.Value[1].Value.Success.Value {
			t.Error("the failed node should decode as a failure")
		}
		if !res.Results.Value[1].Value.Data.Null {
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

		if got := res.Results.Value[0].Value.Message.Value; got != "cleared 1 job(s)" {
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
