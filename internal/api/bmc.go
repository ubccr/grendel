// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"errors"
	"fmt"
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

// BmcPowerBody is one power request, for either a host or its BMC
type BmcPowerBody struct {
	Target       string `json:"target" description:"What to power, 'host' or 'bmc'. Defaults to host. A bmc target is always a graceful restart and ignores the other fields" example:"host"`
	Action       string `json:"action" description:"schemas.ResetType. On, ForceOn, ForceOff, ForceRestart, GracefulRestart, GracefulShutdown, PowerCycle, Nmi, PushPowerButton, Suspend, Pause, Resume, FullPowerCycle" example:"PowerCycle"`
	BootOverride string `json:"boot_override" description:"schemas.BootSource to boot from once. None, Pxe, Floppy, Cd, Usb, Hdd, BiosSetup, Utilities, Diags, UefiShell, UefiTarget, SDCard, UefiHttp, RemoteDrive, UefiBootNext, Recovery" example:"Pxe"`
}

// BmcClearJobsBody names what to clear. Clearing everything is its own field, so no job id is reserved to mean it
type BmcClearJobsBody struct {
	All bool     `json:"all" description:"Clear every job in the queue. Mutually exclusive with ids"`
	IDs []string `json:"ids" oai3:"nullable" description:"Clear only these job ids" example:"JID_000000000001"`
}

// BmcStatus reports one row per node
func (h *Handler) BmcStatus(c fuego.ContextNoBody) (model.BmcResponse[model.RedfishSystem], error) {
	// this does not tag hosts with grendel:serial. Status is a read and changes nothing in the store
	return runBmc(h, c, "bmc status", false,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[model.RedfishSystem] {
			return h.Redfish.Status(ctx, t)
		},
		func(data *model.RedfishSystem) *model.RedfishSystem { return data },
	)
}

// BmcPower changes the power state of node(s), or restarts their BMCs
func (h *Handler) BmcPower(c fuego.ContextWithBody[BmcPowerBody]) (model.BmcResponse[model.BmcAck], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, badRequest(err, "failed to parse the power request")
	}

	req, err := powerRequest(body)
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}

	return runAck(h, c, powerOperation(req), func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.Power(ctx, t, req)
	})
}

// BmcJobs lists the jobs queued on node(s)
func (h *Handler) BmcJobs(c fuego.ContextNoBody) (model.BmcResponse[model.BmcJobList], error) {
	return runBmc(h, c, "bmc jobs", false,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.JobList] {
			return h.Redfish.Jobs(ctx, t)
		},
		jobList,
	)
}

// BmcClearJobs removes jobs from node(s). It is a POST rather than a DELETE with a body, which some intermediaries strip
func (h *Handler) BmcClearJobs(c fuego.ContextWithBody[BmcClearJobsBody]) (model.BmcResponse[model.BmcAck], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, badRequest(err, "failed to parse the clear jobs request")
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

	return runAck(h, c, operation, func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.ClearJobs(ctx, t, req)
	})
}

// BmcSEL reads the system event log of node(s)
func (h *Handler) BmcSEL(c fuego.ContextNoBody) (model.BmcResponse[model.BmcSELLog], error) {
	limit, err := selLimit(c.QueryParam("limit"))
	if err != nil {
		return model.BmcResponse[model.BmcSELLog]{}, err
	}

	return runBmc(h, c, "bmc sel", false,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.SELLog] {
			return h.Redfish.SEL(ctx, t, redfish.SELRequest{Limit: limit})
		},
		selLog,
	)
}

// BmcClearSEL clears the system event log of node(s)
func (h *Handler) BmcClearSEL(c fuego.ContextNoBody) (model.BmcResponse[model.BmcAck], error) {
	return runAck(h, c, "bmc clear sel", func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.ClearSEL(ctx, t)
	})
}

// bmcContext is what the bmc routes need from a request. The ones that take a body and the ones that do not have different fuego context types but the same three methods
type bmcContext interface {
	QueryParam(string) string
	Context() context.Context
	Response() http.ResponseWriter
}

// runAck runs an operation that changes something and returns no data, which is the kind recorded in the event log
func runAck(h *Handler, c bmcContext, operation string, run func(context.Context, []redfish.Target) []redfish.NodeResult[redfish.Ack]) (model.BmcResponse[model.BmcAck], error) {
	return runBmc(h, c, operation, true, run, func(*redfish.Ack) *model.BmcAck { return nil })
}

// runBmc is every bmc route: pick the nodes, contact them, and report. audit is for operations that change something. Reads stay out of the event log, which keeps only a few dozen events and would otherwise lose the power actions to status polling
//
// A completed run always returns 200, since one that found every node unreachable is an accurate report rather than a failed request. The failure count goes in a header so a script can see it without parsing the body
func runBmc[T, D any](
	h *Handler,
	c bmcContext,
	operation string,
	audit bool,
	run func(context.Context, []redfish.Target) []redfish.NodeResult[T],
	data func(*T) *D,
) (model.BmcResponse[D], error) {
	targets, err := h.bmcTargets(c)
	if err != nil {
		return model.BmcResponse[D]{}, err
	}

	// a fan-out over a large nodeset can outlast the write timeout set in server.go, depending on bmc.run_timeout, and would have its response truncated
	if err := http.NewResponseController(c.Response()).SetWriteDeadline(time.Time{}); err != nil {
		log.Debugf("could not clear the write deadline: %s", err)
	}

	runID := ksuid.New().String()
	started := time.Now()
	log.Debugf("%s run=%s nodes=%d", operation, runID, len(targets))

	results := run(c.Context(), targets)

	sort.SliceStable(results, func(i, j int) bool { return natsort.Compare(results[i].Node, results[j].Node) })

	out := make([]model.BmcResult[D], 0, len(results))
	summary := model.BmcSummary{Total: len(results), DurationMs: time.Since(started).Milliseconds()}
	messages := make(model.JobMessageList, 0, len(results))
	var failed []string

	for _, r := range results {
		status := "success"
		if r.Success {
			summary.Succeeded++
		} else {
			summary.Failed++
			status = "error"
			failed = append(failed, r.Node)
		}
		messages = append(messages, model.JobMessage{Status: status, Host: r.Node})

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

	if audit {
		severity := model.SeveritySuccess
		switch {
		case summary.Failed > 0 && summary.Succeeded == 0:
			severity = model.SeverityError
		case summary.Failed > 0:
			severity = model.SeverityWarning
		}

		h.writeEvent(c.Context(), severity.String(),
			fmt.Sprintf("%s: %d succeeded, %d failed of %d (run %s)", operation, summary.Succeeded, summary.Failed, summary.Total, runID),
			messages...)
	}

	return model.BmcResponse[D]{RunID: runID, Summary: summary, Results: out}, nil
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

// bmcTargets resolves a request to the BMCs to contact. A request that resolves to nothing is an error rather than an empty success, since a script has no other way to learn that its nodeset named nothing
func (h *Handler) bmcTargets(c bmcContext) ([]redfish.Target, error) {
	hosts, err := h.bmcSelect(c)
	if err != nil {
		return nil, err
	}
	if len(hosts) == 0 {
		return nil, notFound("no nodes matched the given nodeset and tags")
	}

	targets := redfish.TargetsFromHosts(hosts)
	if len(targets) == 0 {
		return nil, notFound("none of the selected nodes have an interface marked bmc")
	}

	return targets, nil
}

// bmcSelect resolves which nodes a request is about. An empty nodeset does not mean every node, because FindHosts turns an empty slice into `where name in (NULL)`, so all=true is how to ask for that and a request naming no selector at all is rejected. A nodeset or tags given alongside all=true still narrow the selection, because `grendel bmc power off all --tags gpu` must never reach a node without that tag
func (h *Handler) bmcSelect(c bmcContext) (model.HostList, error) {
	nodes := c.QueryParam("nodeset")
	tags := c.QueryParam("tags")
	all := c.QueryParam("all") == "true"

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
		// a manager reset takes no reset type and no boot override, so the rest of the body is not read
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

// powerOperation names a power request for the event log
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

// selLimit caps how much of each log comes back. A populated iDRAC log is thousands of entries, and a nodeset multiplies that, so the default keeps a whole rack readable
func selLimit(raw string) (int, error) {
	if raw == "" {
		return defaultSELLimit, nil
	}

	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, badRequest(err, "limit must be a number")
	}
	if limit < 0 {
		return 0, badRequest(errors.New("negative limit"), "limit must be zero or more, zero returns the whole log")
	}

	return limit, nil
}

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
