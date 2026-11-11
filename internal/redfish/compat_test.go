// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stmcginnis/gofish/schemas"
)

// TestCompat answers one question about one machine: does Grendel's BMC layer work here, and if not, which specific assumption fails. Every heuristic in this package was written against a Dell iDRAC, so pointing it at anything else is the only way to find out what it assumed. See docs/pages/advanced/bmc-compatibility.md for how to capture a mockup off a bmc and serve it
//
// The report is the test output, and it turns on one distinction. A skip means the hardware does not have the thing at all, which is correct behaviour. A failure means the hardware has it and Grendel gets it wrong. Nothing here fails because a bmc is missing a feature
var (
	compatEndpoint    = flag.String("redfish.endpoint", "", "Redfish base url to run the compatibility suite against. Empty skips the suite")
	compatUser        = flag.String("redfish.user", "compat", "bmc username. Any non empty value works against a mockup server, which authenticates nothing")
	compatPassword    = flag.String("redfish.password", "compat", "bmc password")
	compatBasicAuth   = flag.Bool("redfish.basic-auth", false, "log in with basic auth instead of a redfish session. Required against a mockup server, which advertises no session service")
	compatInsecure    = flag.Bool("redfish.insecure", true, "skip tls verification, which every bmc needs since they all ship a self signed cert")
	compatDestructive = flag.Bool("redfish.destructive", false, "run the checks that actually write: power, job clear and sel clear. Off by default because this suite cannot tell a mockup server from a live bmc")
	compatBiosChange  = flag.String("redfish.bios-change", "", "Name=Value of a harmless bios attribute to stage and then discard, such as NumLock=Off. Needs -redfish.destructive. The suite picks none itself, since no attribute is harmless on every machine")
)

// resetTypes the power command tells the operator it accepts, from cmd/grendel/bmc/power.go
var compatAdvertisedResets = []schemas.ResetType{
	schemas.OnResetType, schemas.ForceOnResetType, schemas.ForceOffResetType,
	schemas.ForceRestartResetType, schemas.GracefulRestartResetType, schemas.GracefulShutdownResetType,
	schemas.PowerCycleResetType, schemas.NmiResetType, schemas.PushPowerButtonResetType,
	schemas.SuspendResetType, schemas.PauseResetType, schemas.ResumeResetType, schemas.FullPowerCycleResetType,
}

// the three the power command's shorthands map to. A bmc that does not support these cannot serve `grendel bmc power on|off|cycle` at all, so they are the ones worth failing over
var compatRequiredResets = []schemas.ResetType{
	schemas.OnResetType, schemas.ForceOffResetType, schemas.ForceRestartResetType,
}

// boot sources the power command's --override flag advertises
var compatAdvertisedBoot = []schemas.BootSource{
	schemas.PxeBootSource, schemas.BiosSetupBootSource, schemas.UtilitiesBootSource, schemas.DiagsBootSource,
}

func TestCompat(t *testing.T) {
	if *compatEndpoint == "" {
		t.Skip("set -redfish.endpoint to run the compatibility suite")
	}

	ctx := context.Background()

	c := compatConnect(t, ctx, *compatBasicAuth)
	t.Logf("vendor=%q product=%q redfish=%q root=%q",
		c.service.Vendor, c.service.Product, c.service.RedfishVersion, c.service.ODataID)

	t.Run("auth", func(t *testing.T) { compatAuth(t, ctx) })
	t.Run("status", func(t *testing.T) { compatStatus(t, ctx) })
	t.Run("power", func(t *testing.T) { compatPower(t, ctx) })
	t.Run("jobs", func(t *testing.T) { compatJobs(t, ctx) })
	t.Run("sel", func(t *testing.T) { compatSEL(t, ctx) })
	t.Run("bios", func(t *testing.T) { compatBios(t, ctx) })
	t.Run("identity", func(t *testing.T) { compatIdentity(t, ctx) })
	t.Run("fanout", func(t *testing.T) { compatFanout(t, ctx) })
}

// compatService builds a Service from the flags. Everything the suite does goes through the real configuration path, so a bmc that only works with some setting also tells the operator which setting that is
func compatService(basicAuth bool) *Service {
	return New(Config{
		Credentials: []Credential{{User: *compatUser, Password: *compatPassword}},
		BasicAuth:   basicAuth,
		Insecure:    *compatInsecure,
		Fanout:      1,
		NodeTimeout: 60 * time.Second,

		ClientTimeout:       60 * time.Second,
		DialTimeout:         10 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	})
}

// compatTarget is the single node under test. The endpoint carries a scheme, so it passes through endpoint() untouched and no bmc.scheme is involved
func compatTarget() Target { return Target{Name: "compat", Addr: *compatEndpoint} }

func compatConnect(t *testing.T, ctx context.Context, basicAuth bool) *conn {
	t.Helper()

	c, err := compatService(basicAuth).connect(ctx, compatTarget())
	if err != nil {
		t.Fatalf("cannot reach %s: %s", *compatEndpoint, err)
	}
	t.Cleanup(c.Close)

	return c
}

// compatAuth reports which login modes this service accepts. Neither result is a failure on its own, the point is to record what the operator has to put in grendel.toml
//
// This is the one check that sends anything other than a GET without -redfish.destructive: testing session login means creating a session and then deleting it, which is what every ordinary Grendel run does anyway. Nothing here changes the machine's state
func compatAuth(t *testing.T, ctx context.Context) {
	session := compatTry(ctx, false)
	basic := compatTry(ctx, true)

	t.Logf("redfish session login: %s", compatOutcome(session))
	t.Logf("basic auth login:      %s", compatOutcome(basic))

	if session != nil && basic != nil {
		t.Errorf("neither login mode was accepted, session: %s, basic: %s", session, basic)
	}
}

func compatTry(ctx context.Context, basicAuth bool) error {
	c, err := compatService(basicAuth).connect(ctx, compatTarget())
	if err != nil {
		return err
	}
	c.Close()
	return nil
}

func compatOutcome(err error) string {
	if err == nil {
		return "accepted"
	}
	return "rejected: " + err.Error()
}

// compatStatus checks every field bmc status puts in a column. An empty one is not a cosmetic problem: the operator sees a blank cell and has no way to tell a bmc that does not report the value from a bmc Grendel reads the wrong field on
func compatStatus(t *testing.T, ctx context.Context) {
	if systems, err := compatConnect(t, ctx, *compatBasicAuth).service.Systems(); err == nil && len(systems) == 0 {
		t.Skip("service exposes no computer systems")
	}

	system, err := compatService(*compatBasicAuth).statusOne(ctx, compatTarget())
	if err != nil {
		t.Fatalf("status failed: %s", err)
	}

	fields := []struct {
		name  string
		value any
		empty bool
	}{
		{"host_name", system.HostName, system.HostName == ""},
		{"bios_version", system.BIOSVersion, system.BIOSVersion == ""},
		{"serial_number", system.SerialNumber, system.SerialNumber == ""},
		{"manufacturer", system.Manufacturer, system.Manufacturer == ""},
		{"model", system.Model, system.Model == ""},
		{"power_status", system.PowerStatus, system.PowerStatus == ""},
		{"health", system.Health, system.Health == ""},
		{"total_memory", system.TotalMemory, system.TotalMemory == 0},
		{"processor_count", system.ProcessorCount, system.ProcessorCount == 0},
		{"boot_order", system.BootOrder, len(system.BootOrder) == 0},
	}

	for _, f := range fields {
		t.Run(f.name, func(t *testing.T) {
			if f.empty {
				t.Errorf("bmc status reports an empty %s, so statusOne is reading a field this service does not populate", f.name)
				return
			}
			t.Logf("%v", f.value)
		})
	}

	t.Run("dell_oem", func(t *testing.T) {
		c := compatConnect(t, ctx, *compatBasicAuth)
		if !c.isDell() {
			t.Skipf("service root reports vendor %q, not Dell", c.service.Vendor)
		}

		systems, err := c.service.Systems()
		if err != nil || len(systems) == 0 {
			t.Skip("no computer systems to read an oem block from")
		}

		if _, err := dellOEMSystem(systems[0]); err != nil {
			t.Errorf("service reports Dell but its oem block did not parse: %s", err)
		}
	})
}

// compatPower asks the service what it supports rather than making it do anything. That is where nearly all of the signal is, and it is the only way this can be safe to run against a live bmc
func compatPower(t *testing.T, ctx context.Context) {
	c := compatConnect(t, ctx, *compatBasicAuth)

	systems, err := c.service.Systems()
	if err != nil {
		t.Fatalf("cannot read computer systems: %s", err)
	}
	if len(systems) == 0 {
		t.Skip("service exposes no computer systems")
	}
	sys := systems[0]

	t.Run("reset_types", func(t *testing.T) {
		supported, err := sys.GetSupportedResetTypes()
		if err != nil {
			t.Fatalf("cannot read supported reset types: %s", err)
		}
		if len(supported) == 0 {
			t.Skip("system declares no reset types, neither as AllowableValues nor via ActionInfo")
		}

		t.Logf("supported: %s", compatJoin(supported))
		if missing := compatMissing(compatAdvertisedResets, supported); len(missing) > 0 {
			t.Logf("bmc power advertises but this system does not support: %s", compatJoin(missing))
		}

		if missing := compatMissing(compatRequiredResets, supported); len(missing) > 0 {
			t.Errorf("bmc power on|off|cycle cannot work here, the system does not support %s", compatJoin(missing))
		}
	})

	t.Run("boot_override", func(t *testing.T) {
		allowed := sys.Boot.AllowableBootSourceOverrideTargetValues
		if len(allowed) == 0 {
			t.Skip("system declares no BootSourceOverrideTarget allowable values")
		}

		t.Logf("supported: %s", compatJoin(allowed))
		if missing := compatMissing(compatAdvertisedBoot, allowed); len(missing) > 0 {
			t.Logf("bmc power --override advertises but this system does not support: %s", compatJoin(missing))
		}

		// Pxe is the one provisioning actually needs, the rest are conveniences
		if !slices.Contains(allowed, schemas.PxeBootSource) {
			t.Errorf("this system cannot be told to boot Pxe once, so bmc power --override Pxe cannot provision it")
		}
	})

	t.Run("manager_reset", func(t *testing.T) {
		managers, err := c.service.Managers()
		if err != nil {
			t.Fatalf("cannot read managers: %s", err)
		}
		if len(managers) == 0 {
			t.Skip("service exposes no managers")
		}

		supported, err := managers[0].GetSupportedResetTypes()
		if err != nil {
			t.Fatalf("cannot read supported manager reset types: %s", err)
		}
		if len(supported) == 0 {
			t.Skip("manager declares no reset types")
		}

		t.Logf("supported: %s", compatJoin(supported))
		// resetManagers sends this one unconditionally, so a manager without it cannot serve bmc reboot-bmc
		if !slices.Contains(supported, schemas.GracefulRestartResetType) {
			t.Errorf("resetManagers always sends %s, which this manager does not support", schemas.GracefulRestartResetType)
		}
	})

	t.Run("reset", func(t *testing.T) {
		compatRequireDestructive(t)

		results := compatService(*compatBasicAuth).Power(ctx, []Target{compatTarget()},
			PowerRequest{Target: TargetHost, Reset: schemas.GracefulRestartResetType})
		compatAssertSuccess(t, results[0].Success, results[0].Message)
	})
}

// compatJobs covers the job service. A bmc without one is not broken, which is why absence skips
func compatJobs(t *testing.T, ctx context.Context) {
	c := compatConnect(t, ctx, *compatBasicAuth)

	js, err := c.service.JobService()
	if err != nil {
		if compatAbsent(err) {
			t.Skipf("service does not implement a job service: %s", err)
		}
		t.Fatalf("cannot read the job service: %s", err)
	}
	if js == nil {
		t.Skip("service root declares no JobService")
	}
	t.Logf("job service: %s", js.ODataID)

	// clearing named ids does not use the collection link the service advertises, clearJobsOne builds {JobService}/Jobs/{id} by hand. On a service that puts its jobs anywhere else, every such clear is a DELETE to a 404
	t.Run("delete_uri", func(t *testing.T) {
		var doc struct {
			Jobs struct {
				ODataID string `json:"@odata.id"`
			} `json:"Jobs"`
		}
		if err := compatGetJSON(c, js.ODataID, &doc); err != nil {
			t.Fatalf("cannot re-read the job service: %s", err)
		}
		if doc.Jobs.ODataID == "" {
			t.Skip("job service advertises no Jobs collection link")
		}

		want := js.ODataID + "/Jobs"
		if doc.Jobs.ODataID != want {
			t.Errorf("clearJobsOne builds %s/{id} but this service keeps its jobs at %s, so clearing a named job would DELETE a uri that does not exist", want, doc.Jobs.ODataID)
		}
	})

	jobs, err := js.Jobs()
	if err != nil {
		t.Fatalf("cannot list jobs: %s", err)
	}
	t.Logf("%d job(s) queued", len(jobs))

	t.Run("jobs_parse", func(t *testing.T) {
		if len(jobs) == 0 {
			t.Skip("queue is empty, nothing to parse")
		}

		for _, job := range jobs {
			// bmc job list keys its rows on these two, so a job missing either is a blank row
			if job.ID == "" {
				t.Errorf("job at %s has no Id", job.ODataID)
			}
			if job.JobState == "" {
				t.Errorf("job %s has no JobState", job.ID)
			}

			if want := strings.TrimRight(js.ODataID, "/") + "/Jobs/" + url.PathEscape(job.ID); job.ODataID != want {
				t.Errorf("clearJobsOne would DELETE %s but job %s lives at %s", want, job.ID, job.ODataID)
			}
		}
	})

	t.Run("clear", func(t *testing.T) {
		compatRequireDestructive(t)

		results := compatService(*compatBasicAuth).ClearJobs(ctx, []Target{compatTarget()},
			ClearJobsRequest{All: true})
		compatAssertSuccess(t, results[0].Success, results[0].Message)
	})
}

// compatSEL covers the two heuristics most likely to be wrong on hardware nobody has tried: which log service is the system event log, and what its timestamps look like
func compatSEL(t *testing.T, ctx context.Context) {
	c := compatConnect(t, ctx, *compatBasicAuth)

	matched, err := selLogServices(c)
	if err != nil {
		t.Fatalf("cannot read log services: %s", err)
	}

	for _, l := range matched {
		t.Logf("log service %s: matched, LogEntryType is %q", l.ODataID, l.LogEntryType)
	}

	if len(matched) == 0 {
		t.Skip("no log service on any manager or system looks like a system event log. If this bmc has one, its id belongs in knownSystemLogIDs")
	}
	if len(matched) > 1 {
		// selOne concatenates every match, so bmc sel list here is not only the system event log
		t.Logf("%d log services matched, selOne concatenates all of them into one log", len(matched))
	}

	t.Run("clear_log_declared", func(t *testing.T) {
		// clearSELOne attempts every match and fails the node if any refuses, so one that declares no action makes bmc sel clear fail here
		for _, l := range matched {
			if !l.SupportsClearLog() {
				t.Errorf("log service %s declares no #LogService.ClearLog action, so bmc sel clear would fail on this node", l.ID)
			}
		}
	})

	sel, err := compatService(*compatBasicAuth).selOne(ctx, compatTarget(), SELRequest{})
	if err != nil {
		t.Fatalf("cannot read the system event log: %s", err)
	}
	t.Logf("%d entries", len(sel.Entries))

	t.Run("timestamps", func(t *testing.T) {
		if len(sel.Entries) == 0 {
			t.Skip("log is empty, nothing to time")
		}

		// selOne sorts on this. An entry entryTime cannot read sorts as the oldest, so it silently falls off the end of a --limit and the operator never sees it
		for _, e := range sel.Entries {
			if entryTime(e).IsZero() {
				t.Errorf("entry %s has no timestamp entryTime can read, Created=%q EventTimestamp=%q. selOne sorts it to the bottom of the log", e.ID, e.Created, e.EventTimestamp)
			}
		}
	})

	t.Run("entry_fields", func(t *testing.T) {
		if len(sel.Entries) == 0 {
			t.Skip("log is empty")
		}

		// the columns bmc sel list prints
		for _, e := range sel.Entries {
			if e.ID == "" {
				t.Errorf("an entry has no Id, so bmc sel list cannot name its row")
			}
			if e.Message == "" {
				t.Errorf("entry %s has no Message, which is the only column anyone reads", e.ID)
			}
			if e.Severity == "" {
				t.Errorf("entry %s has no Severity, so bmc sel list cannot colour it", e.ID)
			}
		}
	})

	t.Run("top_is_newest", func(t *testing.T) {
		const n = 10
		if len(sel.Entries) <= n {
			t.Skipf("log has %d entries, a limit of %d cannot leave any out", len(sel.Entries), n)
		}

		// a limit is sent as $top, which returns the first members in the bmc's own order. Compare that against the whole log sorted, which is the true newest
		limited, err := compatService(*compatBasicAuth).selOne(ctx, compatTarget(), SELRequest{Limit: n})
		if err != nil {
			t.Fatalf("cannot read the system event log with a limit: %s", err)
		}

		want := make([]string, 0, n)
		for _, e := range sel.Entries[:n] {
			want = append(want, e.ID)
		}
		got := make([]string, 0, len(limited.Entries))
		for _, e := range limited.Entries {
			got = append(got, e.ID)
		}
		if !slices.Equal(want, got) {
			t.Errorf("this bmc does not list its log newest first, so bmc sel list --limit %d shows %v rather than the newest entries %v. --limit 0 reads the whole log and sorts it", n, got, want)
		}
	})

	t.Run("clear", func(t *testing.T) {
		compatRequireDestructive(t)

		results := compatService(*compatBasicAuth).ClearSEL(ctx, []Target{compatTarget()})
		compatAssertSuccess(t, results[0].Success, results[0].Message)
	})
}

func compatBios(t *testing.T, ctx context.Context) {
	svc := compatService(*compatBasicAuth)

	bios, err := svc.biosOne(ctx, compatTarget(), BiosRequest{})
	if compatAbsent(err) {
		t.Skipf("bmc has no bios resource: %s", err)
	}
	if err != nil {
		t.Fatalf("cannot read the bios: %s", err)
	}
	t.Logf("%d attributes, %d pending, registry %q, apply times %v", len(bios.Attributes), len(bios.Pending), bios.AttributeRegistry, bios.ApplyTimes)

	t.Run("readable", func(t *testing.T) {
		if len(bios.Attributes) == 0 {
			t.Error("the bios resource lists no attributes, so bmc bios get shows nothing")
		}
	})

	t.Run("settings_object", func(t *testing.T) {
		b, err := svc.connectBios(ctx, compatTarget())
		if err != nil {
			t.Fatal(err)
		}
		defer b.Close()

		if b.settings.SettingsObject == "" || b.settings.SettingsObject == b.bios.ODataID {
			t.Skip("no separate settings object, this bmc applies bios changes directly and has nothing to show as pending or discard")
		}
		t.Logf("changes are staged at %s", b.settings.SettingsObject)
		if len(b.settings.SupportedApplyTimes) == 0 {
			t.Log("the bmc declares no apply times, so the editor offers all four and the bmc may refuse some")
		}
	})

	t.Run("registry_reachable", func(t *testing.T) {
		reg, err := svc.biosRegistryOne(ctx, compatTarget())
		if err != nil {
			t.Fatalf("cannot read the attribute registry: %s", err)
		}
		if !reg.Available {
			t.Skipf("the bmc serves no copy of registry %q itself, so the editor falls back to raw names", bios.AttributeRegistry)
		}

		described := map[string]bool{}
		for _, a := range reg.Attributes {
			described[a.Name] = true
		}
		var missing []string
		for name := range bios.Attributes {
			if !described[name] {
				missing = append(missing, name)
			}
		}
		t.Logf("registry %s %s (%s) describes %d attributes", reg.Registry, reg.Version, reg.Language, len(reg.Attributes))
		if len(missing) > 0 {
			slices.Sort(missing)
			t.Logf("%d attributes are not in the registry and show under Other: %v", len(missing), missing)
		}
	})

	t.Run("stage_and_discard", func(t *testing.T) {
		compatRequireDestructive(t)
		if *compatBiosChange == "" {
			t.Skip("pass -redfish.bios-change Name=Value to stage a change and discard it")
		}
		name, value, ok := strings.Cut(*compatBiosChange, "=")
		if !ok {
			t.Fatalf("-redfish.bios-change %q is not Name=Value", *compatBiosChange)
		}

		// no apply time, so the change stays staged rather than becoming a job that a reboot would apply if the discard fails
		staged, err := svc.setBiosOne(ctx, compatTarget(), SetBiosRequest{Attributes: map[string]any{name: value}})
		if err != nil {
			t.Fatalf("cannot stage %s: %s", *compatBiosChange, err)
		}
		t.Log(staged.Note)

		after, err := svc.biosOne(ctx, compatTarget(), BiosRequest{Names: []string{name}})
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := after.Pending[name]; !ok {
			t.Errorf("%s was staged but does not show as pending, so the editor would not show it", name)
		}

		discarded, err := svc.discardBiosOne(ctx, compatTarget())
		if err != nil {
			t.Fatalf("staged %s but could not discard it, clear it on the bmc before its next reboot: %s", name, err)
		}
		t.Log(discarded.Note)
	})
}

// compatFanout runs the public methods once each, for whatever this service has. Everything above calls the per node functions directly, so this is what confirms Run agrees with them rather than turning a working query into a failed row
func compatFanout(t *testing.T, ctx context.Context) {
	c := compatConnect(t, ctx, *compatBasicAuth)
	svc := compatService(*compatBasicAuth)
	targets := []Target{compatTarget()}

	if systems, err := c.service.Systems(); err == nil && len(systems) > 0 {
		status := svc.Status(ctx, targets)
		compatAssertSuccess(t, status[0].Success, status[0].Message)
	}

	if js, err := c.service.JobService(); err == nil && js != nil {
		jobs := svc.Jobs(ctx, targets)
		compatAssertSuccess(t, jobs[0].Success, jobs[0].Message)
	}

	if logs, err := selLogServices(c); err == nil && len(logs) > 0 {
		sel := svc.SEL(ctx, targets, SELRequest{Limit: 10})
		compatAssertSuccess(t, sel[0].Success, sel[0].Message)
	}
}

func compatAssertSuccess(t *testing.T, success bool, message string) {
	t.Helper()

	if !success {
		t.Error(message)
		return
	}
	if message != "" {
		t.Log(message)
	}
}

func compatIdentity(t *testing.T, ctx context.Context) {
	svc := compatService(*compatBasicAuth)
	name := compatTarget().Name

	t.Run("readable", func(t *testing.T) {
		sys, err := svc.statusOne(ctx, compatTarget())
		if err != nil {
			t.Fatalf("cannot read the system: %s", err)
		}
		if sys.Location == nil {
			t.Log("the chassis reports no location, so bmc status shows none until one is set")
		} else {
			t.Logf("asset tag %q, location %+v", sys.AssetTag, *sys.Location)
		}
	})

	t.Run("location_source", func(t *testing.T) {
		res := svc.SetLocation(ctx, []Target{compatTarget()}, LocationRequest{ByNode: map[string]LocationUpdate{name: {}}, DryRun: true})[0]
		if !res.Success {
			t.Fatalf("cannot read the current location: %s", res.Message)
		}
		t.Logf("current location %+v", res.Data.Current)
	})

	// writes back the values already there, which exercises the PATCH without changing anything
	t.Run("roundtrip", func(t *testing.T) {
		compatRequireDestructive(t)

		cur := svc.SetLocation(ctx, []Target{compatTarget()}, LocationRequest{ByNode: map[string]LocationUpdate{name: {}}, DryRun: true})[0]
		if !cur.Success {
			t.Fatal(cur.Message)
		}
		loc := cur.Data.Current
		marker := loc.Rack + "-grendel"
		update := LocationUpdate{Rack: &marker}
		res := svc.SetLocation(ctx, []Target{compatTarget()}, LocationRequest{ByNode: map[string]LocationUpdate{name: update}})[0]
		if !res.Success {
			t.Fatalf("cannot write the rack: %s", res.Message)
		}
		after := svc.SetLocation(ctx, []Target{compatTarget()}, LocationRequest{ByNode: map[string]LocationUpdate{name: {}}, DryRun: true})[0]
		if after.Success && after.Data.Current.Rack != marker {
			t.Errorf("wrote rack %q but the bmc reads back %q", marker, after.Data.Current.Rack)
		}

		restore := svc.SetLocation(ctx, []Target{compatTarget()}, LocationRequest{ByNode: map[string]LocationUpdate{name: {Rack: &loc.Rack}}})[0]
		if !restore.Success {
			t.Errorf("could not restore rack %q: %s", loc.Rack, restore.Message)
		}
	})
}

// compatRequireDestructive keeps the writing checks off a machine somebody cares about. This suite takes a url and cannot tell a mockup server from a live iDRAC, so the operator has to say
func compatRequireDestructive(t *testing.T) {
	t.Helper()

	if !*compatDestructive {
		t.Skip("this check writes to the bmc, pass -redfish.destructive to run it")
	}
}

// compatAbsent reports a bmc answering that it does not have a resource, which is how the suite tells "this bmc does not have the feature" from "this bmc has it and the request failed". The first is a skip, the second is a failure
func compatAbsent(err error) bool {
	var redfishErr *schemas.Error
	if !errors.As(err, &redfishErr) {
		return false
	}

	switch redfishErr.HTTPReturnedStatusCode {
	case http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusNotImplemented:
		return true
	}
	return false
}

func compatGetJSON(c *conn, uri string, into any) error {
	resp, err := c.client.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, into)
}

// compatMissing returns the members of want that supported does not contain
func compatMissing[T comparable](want, supported []T) []T {
	var missing []T
	for _, w := range want {
		if !slices.Contains(supported, w) {
			missing = append(missing, w)
		}
	}
	return missing
}

func compatJoin[T ~string](values []T) string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, string(v))
	}
	return strings.Join(out, ", ")
}
