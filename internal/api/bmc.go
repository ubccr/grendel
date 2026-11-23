// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/facette/natsort"
	"github.com/go-fuego/fuego"
	"github.com/segmentio/ksuid"
	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

type BmcPowerBody struct {
	Target       string `json:"target" description:"What to power, 'host' or 'bmc'. Defaults to host. A bmc target is always a graceful restart and takes no action or boot_override" example:"host"`
	Action       string `json:"action" description:"schemas.ResetType. On, ForceOn, ForceOff, ForceRestart, GracefulRestart, GracefulShutdown, PowerCycle, Nmi, PushPowerButton, Suspend, Pause, Resume, FullPowerCycle" example:"PowerCycle"`
	BootOverride string `json:"boot_override" description:"schemas.BootSource to boot from once. None, Pxe, Floppy, Cd, Usb, Hdd, BiosSetup, Utilities, Diags, UefiShell, UefiTarget, SDCard, UefiHttp, RemoteDrive, UefiBootNext, Recovery" example:"Pxe"`
}

type BmcIndicatorBody struct {
	State string `json:"state" description:"on or off" example:"on"`
}

// Clearing everything is its own field, so no job id is reserved to mean it
type BmcClearJobsBody struct {
	All bool     `json:"all" description:"Clear every job in the queue. Mutually exclusive with ids"`
	IDs []string `json:"ids" oai3:"nullable" description:"Clear only these job ids" example:"JID_000000000001"`
}

func (h *Handler) BmcStatus(c fuego.ContextNoBody) (model.BmcResponse[model.RedfishSystem], error) {
	return runBmc(h, c, "bmc status", nil,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[model.RedfishSystem] {
			return h.Redfish.Status(ctx, t)
		},
		func(data *model.RedfishSystem) *model.RedfishSystem { return data },
	)
}

func (h *Handler) BmcPower(c fuego.ContextWithBody[BmcPowerBody]) (model.BmcResponse[model.BmcAck], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}

	req, err := powerRequest(body)
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}

	action := "bmc.power"
	if req.Target == redfish.TargetBMC {
		action = "bmc.reboot"
	}

	return runAck(h, c, action, powerOperation(req), func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.Power(ctx, t, req)
	})
}

func (h *Handler) BmcIndicator(c fuego.ContextWithBody[BmcIndicatorBody]) (model.BmcResponse[model.BmcAck], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}

	var req redfish.IndicatorRequest
	switch strings.ToLower(body.State) {
	case "on":
		req.On = true
	case "off":
	default:
		return model.BmcResponse[model.BmcAck]{}, badRequest(fmt.Errorf("unknown indicator state %q", body.State), "state must be on or off")
	}

	return runAck(h, c, "bmc.indicator", "bmc indicator "+strings.ToLower(body.State), func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.Indicator(ctx, t, req)
	})
}

func (h *Handler) BmcJobs(c fuego.ContextNoBody) (model.BmcResponse[model.BmcJobList], error) {
	return runBmc(h, c, "bmc jobs", nil,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.JobList] {
			return h.Redfish.Jobs(ctx, t)
		},
		jobList,
	)
}

// A POST rather than a DELETE with a body, which some intermediaries strip
func (h *Handler) BmcClearJobs(c fuego.ContextWithBody[BmcClearJobsBody]) (model.BmcResponse[model.BmcAck], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}

	// rejecting these here means one 400 rather than one failed row per node in the nodeset
	if !body.All && len(body.IDs) == 0 {
		return model.BmcResponse[model.BmcAck]{}, badRequest(errors.New("no jobs named"), "set all to clear the queue, or name the job ids to remove")
	}
	if body.All && len(body.IDs) > 0 {
		return model.BmcResponse[model.BmcAck]{}, badRequest(errors.New("all and ids both given"), "clear the whole queue or name job ids, not both")
	}
	for _, id := range body.IDs {
		if !validJobID(id) {
			return model.BmcResponse[model.BmcAck]{}, badRequest(fmt.Errorf("invalid job id %q", id), "job ids may only contain letters, numbers, and . _ : -")
		}
	}

	req := redfish.ClearJobsRequest{All: body.All, IDs: body.IDs}

	operation := "bmc clear jobs " + strings.Join(body.IDs, ",")
	if body.All {
		operation = "bmc clear all jobs"
	}

	return runAck(h, c, "bmc.jobs.clear", operation, func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.ClearJobs(ctx, t, req)
	})
}

func (h *Handler) BmcSEL(c fuego.ContextNoBody) (model.BmcResponse[model.BmcSELLog], error) {
	limit, err := c.QueryParamIntErr("limit")
	if err != nil || limit < 0 {
		return model.BmcResponse[model.BmcSELLog]{}, badRequest(fmt.Errorf("bad limit %q", c.QueryParam("limit")), "limit must be a number, zero or more. Zero returns the whole log")
	}

	return runBmc(h, c, "bmc sel", nil,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.SELLog] {
			return h.Redfish.SEL(ctx, t, redfish.SELRequest{Limit: limit})
		},
		selLog,
	)
}

func (h *Handler) BmcClearSEL(c fuego.ContextNoBody) (model.BmcResponse[model.BmcAck], error) {
	return runAck(h, c, "bmc.sel.clear", "bmc clear sel", func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.ClearSEL(ctx, t)
	})
}

// bmcContext is what the bmc routes need from a request. The ones that take a body and the ones that do not have different fuego context types but the same methods
type bmcContext interface {
	QueryParam(string) string
	QueryParamBoolErr(string) (bool, error)
	Context() context.Context
	Response() http.ResponseWriter
}

// runAck runs an operation that changes something and returns no data, which is the kind recorded in the event log. action is the event's name for programs, operation its summary for people
func runAck(h *Handler, c bmcContext, action, operation string, run func(context.Context, []redfish.Target) []redfish.NodeResult[redfish.Ack]) (model.BmcResponse[model.BmcAck], error) {
	return runBmc(h, c, operation, trackBmc[redfish.Ack](true, action), run, func(*redfish.Ack) *model.BmcAck { return nil })
}

// bmcTrack is how a bmc operation is recorded as an event. node, when set, adjusts a node's row from its result, which is how an operation that only queued work leaves its nodes running
type bmcTrack[T any] struct {
	kind   model.EventKind
	action string
	// detail is added to the event's detail, beside the run id
	detail map[string]string
	node   func(row *model.EventNode, success bool, data *T)
}

// runBmc picks the nodes, contacts them and reports. Only operations that change something pass track, so reads stay out of the event log, which status polling would otherwise fill
func runBmc[T, D any](
	h *Handler,
	c bmcContext,
	operation string,
	track *bmcTrack[T],
	run func(context.Context, []redfish.Target) []redfish.NodeResult[T],
	data func(*T) *D,
) (model.BmcResponse[D], error) {
	nodes, err := h.bmcNodes(c)
	if err != nil {
		return model.BmcResponse[D]{}, err
	}

	return runBmcOn(h, c, nodes, operation, track, run, data)
}

// runBmcOn is runBmc for a handler that needed the nodes first, to build its request from them
//
// A completed run always returns 200, since one that found every node unreachable is an accurate report rather than a failed request. The failure count goes in a header so a script can see it without parsing the body
func runBmcOn[T, D any](
	h *Handler,
	c bmcContext,
	selected bmcNodes,
	operation string,
	track *bmcTrack[T],
	run func(context.Context, []redfish.Target) []redfish.NodeResult[T],
	data func(*T) *D,
) (model.BmcResponse[D], error) {
	targets := redfish.TargetsFromHosts(selected.hosts)

	// a fan-out over a large nodeset can outlast the write timeout set in server.go, depending on bmc.run_timeout, and would have its response truncated
	if err := http.NewResponseController(c.Response()).SetWriteDeadline(time.Time{}); err != nil {
		log.Debugf("could not clear the write deadline: %s", err)
	}

	runID := ksuid.New().String()
	started := time.Now()
	log.Debugf("%s run=%s nodes=%d", operation, runID, len(targets))

	// the event is written before the BMCs are contacted, so a long run shows while it is going. A run that leaves its event open, as queued firmware updates do, has it closed by whatever follows it
	var event *model.Event
	if track != nil {
		pending := make(model.EventNodeList, 0, len(targets)+len(selected.noBMC))
		for _, t := range targets {
			pending = append(pending, model.EventNode{Node: t.Name, State: model.EventNodePending})
		}
		for _, name := range selected.noBMC {
			pending = append(pending, model.EventNode{Node: name, State: model.EventNodePending})
		}
		detail := map[string]string{"run": runID}
		maps.Copy(detail, track.detail)
		event = h.Events.Open(model.Event{
			Kind:    track.kind,
			Action:  track.action,
			Summary: operation,
			Actor:   actor(c.Context()),
			Detail:  detail,
			Nodes:   pending,
		})
	}

	results := run(c.Context(), targets)
	for _, name := range selected.noBMC {
		results = append(results, redfish.NodeResult[T]{Node: name, Message: "node has no interface marked bmc"})
	}

	sort.SliceStable(results, func(i, j int) bool { return natsort.Compare(results[i].Node, results[j].Node) })

	out := make([]model.BmcResult[D], 0, len(results))
	summary := model.BmcSummary{Total: len(results), DurationMs: time.Since(started).Milliseconds()}
	nodes := make([]model.EventNode, 0, len(results))
	var failed []string

	for _, r := range results {
		state := model.EventNodeSucceeded
		if r.Success {
			summary.Succeeded++
		} else {
			summary.Failed++
			state = model.EventNodeFailed
			failed = append(failed, r.Node)
		}
		// redfish.Run reports on every target, so no node is left pending
		node := model.EventNode{Node: r.Node, State: state, Message: r.Message}
		if track != nil && track.node != nil {
			track.node(&node, r.Success, r.Data)
		}
		nodes = append(nodes, node)

		out = append(out, model.BmcResult[D]{
			Node:       r.Node,
			Success:    r.Success,
			Message:    r.Message,
			DurationMs: r.Duration.Milliseconds(),
			Data:       data(r.Data),
		})
	}
	summary.FailedNodes = foldNodes(failed)

	c.Response().Header().Set("X-Grendel-Failed", strconv.Itoa(summary.Failed))

	if event != nil {
		h.Events.SetNodes(event, nodes...)
		h.Events.Settle(event)
	}

	return model.BmcResponse[D]{RunID: runID, Summary: summary, Results: out}, nil
}

// trackBmc records a bmc operation as an event when on, which a dry run is not
func trackBmc[T any](on bool, action string) *bmcTrack[T] {
	if !on {
		return nil
	}
	return &bmcTrack[T]{kind: model.EventKindBmc, action: action}
}

// foldNodes joins names into a nodeset string, or a plain list if they do not fold
func foldNodes(names []string) string {
	joined := strings.Join(names, ",")
	ns, err := nodeset.NewNodeSet(joined)
	if err != nil {
		return joined
	}
	return ns.String()
}

type bmcNodes struct {
	// hosts are the nodes to contact, each with an interface marked bmc, in natural name order
	hosts model.HostList
	// noBMC are the nodes the nodeset named that have no interface marked bmc, which are reported as failed. A node without one that tags or all=true swept in is left out instead
	noBMC []string
}

// bmcNodes resolves a request to the nodes it acts on. A request that resolves to nothing is an error rather than an empty success, since a script has no other way to learn that its nodeset named nothing
func (h *Handler) bmcNodes(c bmcContext) (bmcNodes, error) {
	selected, err := h.bmcSelect(c)
	if err != nil {
		return bmcNodes{}, err
	}
	if len(selected) == 0 {
		return bmcNodes{}, notFound("no nodes matched the given nodeset and tags")
	}

	named := c.QueryParam("nodeset") != ""
	var nodes bmcNodes
	for _, host := range selected {
		switch {
		case host == nil:
		case host.InterfaceBMC() != nil:
			nodes.hosts = append(nodes.hosts, host)
		case named:
			nodes.noBMC = append(nodes.noBMC, host.Name)
		}
	}
	if len(nodes.hosts) == 0 {
		return bmcNodes{}, notFound("none of the selected nodes have an interface marked bmc")
	}
	sort.SliceStable(nodes.hosts, func(i, j int) bool { return natsort.Compare(nodes.hosts[i].Name, nodes.hosts[j].Name) })

	return nodes, nil
}

// bmcSelect resolves which nodes a request is about. An empty nodeset does not mean every node, because FindHosts turns an empty slice into `where name in (NULL)`, so all=true is how to ask for that and a request naming no selector at all is rejected. A nodeset or tags given alongside all=true still narrow the selection, because `grendel bmc power off all --tags gpu` must never reach a node without that tag
func (h *Handler) bmcSelect(c bmcContext) (model.HostList, error) {
	nodes := c.QueryParam("nodeset")
	tags := c.QueryParam("tags")
	all, err := c.QueryParamBoolErr("all")
	if err != nil {
		return nil, badRequest(err, "all must be true or false")
	}

	if !all && nodes == "" && tags == "" {
		return nil, badRequest(errors.New("no node selector given"), "specify nodeset, tags, or all=true")
	}

	if all && nodes == "" && tags == "" {
		hosts, err := h.DB.Hosts()
		if err != nil {
			return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to list nodes"}
		}
		return hosts, nil
	}

	if nodes != "" {
		hosts, err := h.requireNodes(nodes)
		if err != nil || tags == "" {
			return hosts, err
		}
	}

	ns, err := h.filterByNodesetAndTags(nodes, tags)
	if err != nil {
		return nil, badRequest(err, "failed to filter nodes")
	}

	hosts, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to find nodes"}
	}

	return hosts, nil
}

// requireNodes loads every node a nodeset names, and fails the request if any is not in the database, before a single BMC is contacted. A mistyped name usually means the command is not what the operator meant, so acting on the rest of the nodeset is the wrong response when it can power machines off. This looks at the nodeset as given, because a node that exists but lacks a requested tag is filtered out rather than mistyped
func (h *Handler) requireNodes(nodes string) (model.HostList, error) {
	ns, err := nodeset.NewNodeSet(nodes)
	if err != nil {
		return nil, badRequest(err, "failed to parse nodeset")
	}

	hosts, err := h.DB.FindHosts(ns)
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to find nodes"}
	}

	found := make(map[string]bool, len(hosts))
	for _, host := range hosts {
		found[host.Name] = true
	}

	var missing []string
	for _, name := range ns.Iterator().StringSlice() {
		if !found[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, notFound(fmt.Sprintf("nodes not found, no action taken: %s", foldNodes(missing)))
	}

	return hosts, nil
}

// powerRequest validates the body before any node is contacted. A misspelled reset type is one 400 rather than several hundred identical redfish errors
func powerRequest(body BmcPowerBody) (redfish.PowerRequest, error) {
	req := redfish.PowerRequest{Target: redfish.TargetHost}

	switch body.Target {
	case "", string(redfish.TargetHost):
	case string(redfish.TargetBMC):
		// a manager reset is always graceful, so a body asking for anything else is refused rather than answered with a restart it did not ask for
		if body.Action != "" || body.BootOverride != "" {
			return req, badRequest(errors.New("action or boot_override given with a bmc target"), "a bmc target takes no action or boot_override")
		}
		return redfish.PowerRequest{Target: redfish.TargetBMC}, nil
	default:
		return req, badRequest(fmt.Errorf("unknown power target %q", body.Target), "target must be host or bmc")
	}

	if body.Action == "" {
		return req, badRequest(errors.New("no action given"), "action is required when powering a host")
	}
	if !slices.Contains(resetTypes, schemas.ResetType(body.Action)) {
		return req, badRequest(fmt.Errorf("unknown reset type %q", body.Action), "action must be a redfish ResetType")
	}
	req.Reset = schemas.ResetType(body.Action)

	if body.BootOverride != "" {
		if !slices.Contains(bootSources, schemas.BootSource(body.BootOverride)) {
			return req, badRequest(fmt.Errorf("unknown boot source %q", body.BootOverride), "boot_override must be a redfish BootSource")
		}
		req.Boot = schemas.BootSource(body.BootOverride)
	}

	return req, nil
}

func powerOperation(req redfish.PowerRequest) string {
	if req.Target == redfish.TargetBMC {
		return "bmc reboot"
	}
	if req.Boot != "" {
		return fmt.Sprintf("bmc power %s, boot override %s", req.Reset, req.Boot)
	}
	return fmt.Sprintf("bmc power %s", req.Reset)
}

var resetTypes = []schemas.ResetType{
	schemas.OnResetType, schemas.ForceOnResetType, schemas.ForceOffResetType,
	schemas.ForceRestartResetType, schemas.GracefulRestartResetType, schemas.GracefulShutdownResetType,
	schemas.PowerCycleResetType, schemas.NmiResetType, schemas.PushPowerButtonResetType,
	schemas.SuspendResetType, schemas.PauseResetType, schemas.ResumeResetType,
	schemas.FullPowerCycleResetType,
}

var bootSources = []schemas.BootSource{
	schemas.NoneBootSource, schemas.PxeBootSource, schemas.FloppyBootSource, schemas.CdBootSource,
	schemas.UsbBootSource, schemas.HddBootSource, schemas.BiosSetupBootSource, schemas.UtilitiesBootSource,
	schemas.DiagsBootSource, schemas.UefiShellBootSource, schemas.UefiTargetBootSource, schemas.SDCardBootSource,
	schemas.UefiHTTPBootSource, schemas.RemoteDriveBootSource, schemas.UefiBootNextBootSource, schemas.RecoveryBootSource,
}

var jobIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]+$`)

// validJobID keeps a job id to a single path segment, since it becomes part of a DELETE sent with the BMC's admin credentials
func validJobID(id string) bool {
	return jobIDPattern.MatchString(id) && id != "." && id != ".."
}

// defaultSELLimit keeps a whole rack readable: a populated iDRAC log is thousands of entries, and a nodeset multiplies that
const defaultSELLimit = 50

// selLog narrows gofish's log entry the same way jobList narrows its job
func selLog(data *redfish.SELLog) *model.BmcSELLog {
	if data == nil {
		return nil
	}

	out := &model.BmcSELLog{Entries: make([]model.BmcSELEntry, 0, len(data.Entries))}
	for _, e := range data.Entries {
		if e == nil {
			continue
		}

		out.Entries = append(out.Entries, model.BmcSELEntry{
			ID:         e.ID,
			Created:    e.Created,
			Severity:   string(e.Severity),
			Message:    e.Message,
			MessageID:  e.MessageID,
			EntryType:  string(e.EntryType),
			SensorType: string(e.SensorType),
			Resolution: e.Resolution,
		})
	}

	return out
}

// jobList narrows gofish's job to the fields worth sending, see model.BmcJob
func jobList(data *redfish.JobList) *model.BmcJobList {
	if data == nil {
		return nil
	}

	out := &model.BmcJobList{Jobs: make([]model.BmcJob, 0, len(data.Jobs))}
	for _, job := range data.Jobs {
		if job == nil {
			continue
		}

		out.Jobs = append(out.Jobs, model.BmcJob{
			ID:              job.ID,
			Name:            job.Name,
			State:           string(job.JobState),
			Status:          string(job.JobStatus),
			Type:            string(job.JobType),
			PercentComplete: percent(job.PercentComplete),
			StartTime:       job.StartTime,
			EndTime:         job.EndTime,
			Message:         jobMessage(job.Messages),
		})
	}

	return out
}

func percent(p *uint) int {
	if p == nil {
		return 0
	}
	return int(*p)
}

// jobMessage joins what the job has to say into one field, since a job that failed usually explains itself across several messages
func jobMessage(messages []schemas.Message) string {
	var out string
	for _, m := range messages {
		if m.Message == "" {
			continue
		}
		if out != "" {
			out += "; "
		}
		out += m.Message
	}
	return out
}

func badRequest(err error, detail string) fuego.HTTPError {
	return fuego.HTTPError{
		Status: http.StatusBadRequest,
		Err:    err,
		Title:  "Error",
		Detail: detail,
	}
}

func notFound(detail string) fuego.HTTPError {
	return fuego.HTTPError{
		Status: http.StatusNotFound,
		Err:    errors.New(detail),
		Title:  "Error",
		Detail: detail,
	}
}
