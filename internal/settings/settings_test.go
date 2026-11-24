// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import (
	"github.com/pelletier/go-toml/v2"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memBackend struct {
	rows     map[string]Row
	revision int64
}

func (m *memBackend) Settings() ([]Row, int64, error) {
	out := make([]Row, 0, len(m.rows))
	for _, row := range m.rows {
		out = append(out, row)
	}
	return out, m.revision, nil
}

// userRows are the rows someone stored, without the secrets generated at load
func (m *memBackend) userRows() map[string]Row {
	out := map[string]Row{}
	for key, row := range m.rows {
		if o, ok := registry[key]; !ok || !o.Generated {
			out[key] = row
		}
	}
	return out
}

func (m *memBackend) SettingsRevision() (int64, error) { return m.revision, nil }

func (m *memBackend) InsertSettings(set []Row, actor string) (int64, error) {
	for _, row := range set {
		if _, ok := m.rows[row.Key]; !ok {
			m.rows[row.Key] = row
		}
	}
	m.revision++
	return m.revision, nil
}

func (m *memBackend) UpdateSettings(set []Row, unset []string, actor string) (int64, error) {
	for _, row := range set {
		m.rows[row.Key] = row
	}
	for _, key := range unset {
		delete(m.rows, key)
	}
	m.revision++
	return m.revision, nil
}

// start resolves the settings as a fresh process would, with external standing in for the flags, environment and config file
func start(t *testing.T, b *memBackend, external map[string]any, src Source) []Problem {
	t.Helper()
	Reset()
	t.Cleanup(Reset)

	Configure(Env{
		Persistent: true,
		Lookup: func(o *Option) (any, Source, bool) {
			v, ok := external[o.Key]
			return v, src, ok
		},
	})
	warnings, err := Load(b)
	require.NoError(t, err)
	return warnings
}

func find(t *testing.T, key string) View {
	t.Helper()
	for _, section := range List().Sections {
		for _, o := range section.Options {
			if o.Key == key {
				return o
			}
		}
	}
	t.Fatalf("no setting %s", key)
	return View{}
}

func TestCoerce(t *testing.T) {
	tests := []struct {
		opt     *Option
		raw     any
		want    any
		wantErr bool
	}{
		{DNSTTL.Option, int64(600), 600, false},
		{DNSTTL.Option, float64(600), 600, false},
		{DNSTTL.Option, " 600 ", 600, false},
		{DNSTTL.Option, "-1", nil, true},
		{DNSTTL.Option, "soon", nil, true},
		{DNSTTL.Option, 1.5, nil, true},
		{DNSQueryTimeout.Option, "5s", 5 * time.Second, false},
		{DNSQueryTimeout.Option, "0", time.Duration(0), false},
		{DNSQueryTimeout.Option, "60d", nil, true},
		{DNSQueryTimeout.Option, int64(5), nil, true},
		{DNSForward.Option, "", "", false},
		{DNSForward.Option, "1.1.1.1:53", "1.1.1.1:53", false},
		{DNSForward.Option, "1.1.1.1", nil, true},
		{DNSListen.Option, "", nil, true},
	}
	for _, tt := range tests {
		got, err := tt.opt.Coerce(tt.raw)
		if tt.wantErr {
			assert.Error(t, err, "%s <- %v", tt.opt.Key, tt.raw)
			continue
		}
		assert.NoError(t, err, "%s <- %v", tt.opt.Key, tt.raw)
		assert.Equal(t, tt.want, got, "%s <- %v", tt.opt.Key, tt.raw)
	}
}

func TestText(t *testing.T) {
	assert.Equal(t, "2160h", Text(2160*time.Hour))
	assert.Equal(t, "1h30m", Text(90*time.Minute))
	assert.Equal(t, "2s", Text(2*time.Second))
	assert.Equal(t, "300", Text(300))
}

func TestDefaultsWhenNothingIsSet(t *testing.T) {
	start(t, &memBackend{rows: map[string]Row{}}, nil, "")

	assert.Equal(t, 300, DNSTTL.Get())
	assert.Equal(t, SourceDefault, find(t, "dns.ttl").Source)
}

func TestDatabaseValueWinsOverADefault(t *testing.T) {
	b := &memBackend{rows: map[string]Row{"dns.ttl": {Key: "dns.ttl", Value: []byte("600")}}}
	start(t, b, nil, "")

	assert.Equal(t, 600, DNSTTL.Get())
	assert.Equal(t, SourceDatabase, find(t, "dns.ttl").Source)
}

func TestFileWinsOverTheDatabase(t *testing.T) {
	b := &memBackend{rows: map[string]Row{"dns.ttl": {Key: "dns.ttl", Value: []byte("600")}}}
	start(t, b, map[string]any{"dns.ttl": int64(900)}, SourceFile)

	assert.Equal(t, 900, DNSTTL.Get())
	got := find(t, "dns.ttl")
	assert.Equal(t, SourceFile, got.Source)
	assert.Equal(t, SourceFile, got.Source)
	assert.True(t, got.Shadowed)
	assert.Equal(t, 600, got.Stored)

	refused, err := Update(b, map[string]any{"dns.ttl": "30"}, nil, "tester")
	require.NoError(t, err)
	assert.Contains(t, refused["dns.ttl"], "config file")
}

func TestBadValuesAreSkippedNotFatal(t *testing.T) {
	b := &memBackend{rows: map[string]Row{
		"dns.ttl":      {Key: "dns.ttl", Value: []byte(`"soon"`)},
		"bmc.password": {Key: "bmc.password", Value: []byte(`"leaked"`)},
		"no.such.key":  {Key: "no.such.key", Value: []byte(`1`)},
	}}
	warnings := start(t, b, map[string]any{"dns.query_timeout": "60d"}, SourceFile)

	assert.Len(t, warnings, 4)
	assert.Equal(t, 300, DNSTTL.Get())
	assert.Equal(t, 2*time.Second, DNSQueryTimeout.Get())
	assert.Equal(t, SourceDefault, find(t, "bmc.password").Source)
}

func TestLiveOptionAppliesAtOnceAndRestartOptionIsPending(t *testing.T) {
	b := &memBackend{rows: map[string]Row{}}
	start(t, b, nil, "")

	refused, err := Update(b, map[string]any{"dns.forward": "1.1.1.1:53", "dns.ttl": "600"}, nil, "tester")
	require.NoError(t, err)
	require.Empty(t, refused)

	assert.Equal(t, "1.1.1.1:53", DNSForward.Get())
	assert.False(t, find(t, "dns.forward").Pending)

	assert.Equal(t, 300, DNSTTL.Get())
	ttl := find(t, "dns.ttl")
	assert.True(t, ttl.Pending)
	assert.Equal(t, 300, ttl.Running)
	assert.Equal(t, 600, ttl.Value)
	assert.Equal(t, 1, List().Pending)
	assert.JSONEq(t, "600", string(b.rows["dns.ttl"].Value))

	_, err = Update(b, nil, []string{"dns.ttl"}, "tester")
	require.NoError(t, err)
	assert.False(t, find(t, "dns.ttl").Pending)
	assert.NotContains(t, b.rows, "dns.ttl")
}

func TestOneBadChangeSavesNothing(t *testing.T) {
	b := &memBackend{rows: map[string]Row{}}
	start(t, b, nil, "")

	refused, err := Update(b, map[string]any{
		"dns.forward":        "9.9.9.9:53",
		"dns.ttl":            "-5",
		"bmc.password":       "x",
		"provision.repo_dir": "/etc",
		"dbpath":             "/tmp/x.db",
		"nope":               "1",
	}, nil, "tester")
	require.NoError(t, err)

	assert.Len(t, refused, 5)
	assert.NotContains(t, refused, "dns.forward")
	assert.Empty(t, b.userRows())
	assert.Equal(t, "", DNSForward.Get())
}

func TestSecretsNeverCarryAValue(t *testing.T) {
	start(t, &memBackend{rows: map[string]Row{}}, map[string]any{"bmc.password": "hunter2"}, SourceFile)

	got := find(t, "bmc.password")
	assert.Equal(t, SourceFile, got.Source)
	assert.Nil(t, got.Value)
	assert.Nil(t, got.Running)
	assert.Nil(t, got.Stored)
	assert.Nil(t, got.Default)
}

func TestListsAcceptEveryWayTheyAreWritten(t *testing.T) {
	for _, raw := range []any{
		[]any{"10.0.0.1", "10.0.0.2"},
		[]string{"10.0.0.1", "10.0.0.2"},
		`["10.0.0.1","10.0.0.2"]`,
		"[10.0.0.1,10.0.0.2]",
		"10.0.0.1, 10.0.0.2",
		"10.0.0.1 10.0.0.2",
	} {
		got, err := DHCPDNSServers.Coerce(raw)
		assert.NoError(t, err, "%v", raw)
		assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, got, "%v", raw)
	}

	_, err := DHCPDNSServers.Coerce("10.0.0.1,not-an-ip")
	assert.ErrorContains(t, err, "not-an-ip")

	got, err := BMCIPMICipherSuites.Coerce([]any{int64(17), "3"})
	assert.NoError(t, err)
	assert.Equal(t, []int{17, 3}, got)
}

func TestEnumAndBool(t *testing.T) {
	got, err := BMCScheme.Coerce("HTTP")
	assert.NoError(t, err)
	assert.Equal(t, "http", got)
	_, err = BMCScheme.Coerce("ftp")
	assert.ErrorContains(t, err, "https, http")

	for raw, want := range map[any]bool{true: true, "true": true, "false": false, "1": true} {
		got, err := APIWebUI.Coerce(raw)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	}
	_, err = APIWebUI.Coerce("yes please")
	assert.Error(t, err)
}

func TestSubnetsAreValidated(t *testing.T) {
	// a config file writes mtu as text in some installs and its lists as one line of text, and viper lowercases the column names
	got, err := DHCPSubnets.Coerce(map[string]any{"Rack-F": map[string]any{"gateway": "10.17.41.254/23", "dns": "10.17.40.248, 10.17.40.249", "mtu": "9000"}})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"rack-f": map[string]any{"gateway": "10.17.41.254/23", "dns": []string{"10.17.40.248", "10.17.40.249"}, "mtu": 9000}}, got)

	bad := func(record map[string]any) error {
		_, err := DHCPSubnets.Coerce(map[string]any{"a": record})
		return err
	}
	assert.ErrorContains(t, bad(map[string]any{"gateway": "10.17.41.254"}), "prefix length")
	assert.ErrorContains(t, bad(map[string]any{"gateway": "10.17.41.254/23", "dns": "nope"}), "nope")
	assert.ErrorContains(t, bad(map[string]any{"gateway": "10.17.41.254/23", "color": "blue"}), "color")
	assert.ErrorContains(t, bad(map[string]any{"mtu": 9000}), "gateway must be set")
	_, err = DHCPSubnets.Coerce(map[string]any{"rack f": map[string]any{"gateway": "10.17.41.254/23"}})
	assert.ErrorContains(t, err, "not a name")
}

// subnets had no names before, and a config file that still lists them keeps working, in the order it lists them
func TestUnnamedSubnetsAreNumberedByPosition(t *testing.T) {
	got, err := DHCPSubnets.Coerce([]any{
		map[string]any{"gateway": "10.1.0.254/24"},
		map[string]any{"gateway": "10.2.0.254/24"},
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"10": map[string]any{"gateway": "10.1.0.254/24"},
		"20": map[string]any{"gateway": "10.2.0.254/24"},
	}, got)
}

func TestSubnetsFromTheFileAndTheDatabaseSitSideBySide(t *testing.T) {
	b := &memBackend{rows: map[string]Row{}}
	start(t, b, map[string]any{"dhcp.subnets": []any{map[string]any{"gateway": "10.1.0.254/24"}}}, SourceFile)

	refused, err := Update(b, map[string]any{
		"dhcp.subnets.15-lab.gateway": "10.9.0.254/24",
		"dhcp.subnets.15-lab.mtu":     "9000",
		"dhcp.subnets.15-lab.dns":     []any{"10.9.0.1"},
	}, nil, "tester")
	require.NoError(t, err)
	require.Empty(t, refused)
	assert.Equal(t, map[string]any{
		"10":     map[string]any{"gateway": "10.1.0.254/24"},
		"15-lab": map[string]any{"gateway": "10.9.0.254/24", "mtu": 9000, "dns": []string{"10.9.0.1"}},
	}, DHCPSubnets.Get())
	assert.JSONEq(t, `9000`, string(b.rows["dhcp.subnets.15-lab.mtu"].Value), "each field of an entry has a row of its own")
	assert.JSONEq(t, `["10.9.0.1"]`, string(b.rows["dhcp.subnets.15-lab.dns"].Value))
	assert.NotContains(t, b.rows, "dhcp.subnets")

	got := find(t, "dhcp.subnets")
	assert.Equal(t, []string{"10"}, got.FileKeys)

	// each field is listed as a setting of its own, in the order the subnets are tried
	mtu := find(t, "dhcp.subnets.15-lab.mtu")
	assert.Equal(t, 9000, mtu.Value)
	assert.Equal(t, SourceDatabase, mtu.Source)
	assert.Equal(t, "Subnet 15-lab", mtu.Group)
	assert.Equal(t, "dhcp.subnets", mtu.EntryOf)
	assert.Equal(t, 9000, mtu.Stored)
	fromFile := find(t, "dhcp.subnets.10.gateway")
	assert.Equal(t, SourceFile, fromFile.Source)
	assert.Equal(t, SourceFile, find(t, "dhcp.subnets.10.mtu").Source, "a field the file leaves out of its subnet is still the file's")
	assert.Equal(t, []string{"10", "15-lab"}, got.Value)

	_, err = Update(b, map[string]any{"dhcp.subnets.15-lab.mtu": ""}, nil, "tester")
	require.NoError(t, err)
	assert.Equal(t, SourceDefault, find(t, "dhcp.subnets.15-lab.mtu").Source)
	assert.NotContains(t, b.rows, "dhcp.subnets.15-lab.mtu")
	_, err = Update(b, map[string]any{"dhcp.subnets.15-lab.mtu": "9000"}, nil, "tester")
	require.NoError(t, err)

	refused, err = Update(b, map[string]any{"dhcp.subnets.10.mtu": "1400"}, nil, "tester")
	require.NoError(t, err)
	assert.Contains(t, refused["dhcp.subnets.10.mtu"], "config file")

	refused, err = Update(b, map[string]any{"dhcp.subnets.20-new.mtu": "1400"}, nil, "tester")
	require.NoError(t, err)
	assert.Equal(t, "must be set", refused["dhcp.subnets.20-new.gateway"], "an entry with no gateway is refused, against the field that is missing")
	refused, err = Update(b, map[string]any{"dhcp.subnets": `{"a":{"gateway":"10.3.0.254/24"}}`}, nil, "tester")
	require.NoError(t, err)
	assert.Contains(t, refused["dhcp.subnets"], "one field at a time")

	refused, err = Update(b, map[string]any{"dhcp.subnets.15-lab.color": "blue"}, nil, "tester")
	require.NoError(t, err)
	assert.Contains(t, refused["dhcp.subnets.15-lab.color"], "name one of its fields")

	_, err = Update(b, nil, []string{"dhcp.subnets.15-lab.mtu"}, "tester")
	require.NoError(t, err)
	assert.NotContains(t, b.rows, "dhcp.subnets.15-lab.mtu")
	assert.Contains(t, b.rows, "dhcp.subnets.15-lab.gateway")

	_, err = Update(b, nil, []string{"dhcp.subnets.15-lab"}, "tester")
	require.NoError(t, err)
	assert.Empty(t, b.userRows(), "an entry goes with every row under it")
	refused, err = Update(b, nil, []string{"dhcp.subnets.15-lab"}, "tester")
	require.NoError(t, err)
	assert.Contains(t, refused["dhcp.subnets.15-lab"], "no such entry")
}

func TestMapEntriesMergeAndTheFileWins(t *testing.T) {
	b := &memBackend{rows: map[string]Row{
		"provision.extra_vars": {Key: "provision.extra_vars", Value: []byte(`{"timezone":"UTC","root_password":"from-db"}`)},
	}}
	start(t, b, map[string]any{"provision.extra_vars": map[string]any{"root_password": "from-file"}}, SourceFile)

	assert.Equal(t, map[string]any{"timezone": "UTC", "root_password": "from-file"}, ProvisionExtraVars.Get())

	got := find(t, "provision.extra_vars")
	assert.Equal(t, []string{"root_password"}, got.FileKeys)
	assert.NotContains(t, got.Value, "from-file", "an entry from the file may be a secret")
	assert.Equal(t, map[string]any{"timezone": "UTC"}, got.Value)

	refused, err := Update(b, map[string]any{"provision.extra_vars": `{"timezone":"EST","ntp":["a","b"]}`}, nil, "tester")
	require.NoError(t, err)
	require.Empty(t, refused)
	assert.Equal(t, map[string]any{"timezone": "EST", "ntp": []any{"a", "b"}, "root_password": "from-file"}, ProvisionExtraVars.Get())
}

func TestLoggersAreOnUnlessTurnedOff(t *testing.T) {
	b := &memBackend{rows: map[string]Row{"loggers.dhcp": {Key: "loggers.dhcp", Value: []byte(`"off"`)}}}
	start(t, b, map[string]any{"loggers.dns": "OFF"}, SourceFile)

	assert.Equal(t, "off", LoggerLevel("DHCP"))
	assert.Equal(t, "off", LoggerLevel("dns"))
	assert.Equal(t, "warn", LoggerLevel("API"))
	assert.Equal(t, "warn", LoggerLevel("a logger with no setting"))
	assert.Equal(t, "warn", MostVerboseLogLevel())

	_, err := Update(b, map[string]any{"log_level": "info", "loggers.tftp": "debug", "loggers.pxe": "warn"}, nil, "tester")
	require.NoError(t, err)
	assert.Equal(t, "info", LoggerLevel("API"), "a logger left on follows log_level")
	assert.Equal(t, "debug", LoggerLevel("tftp"))
	assert.Equal(t, "warn", LoggerLevel("pxe"), "a logger can be quieter than log_level")
	assert.Equal(t, "off", LoggerLevel("DHCP"))
	assert.Equal(t, "debug", MostVerboseLogLevel())
}

func TestPerNOSSettingFollowsTheSharedOneUntilSet(t *testing.T) {
	b := &memBackend{rows: map[string]Row{}}
	start(t, b, nil, "")

	assert.Equal(t, 6030, SwitchGNMISonic.Port.Get())
	got := find(t, "switch.gnmi.sonic.port")
	assert.Equal(t, SourceInherited, got.Source)
	assert.Equal(t, "switch.gnmi.port", got.InheritFrom)

	_, err := Update(b, map[string]any{"switch.gnmi.port": "7000", "switch.gnmi.plaintext": "true"}, nil, "tester")
	require.NoError(t, err)
	assert.Equal(t, 7000, SwitchGNMISonic.Port.Get(), "a change to the shared setting reaches the ones following it")
	assert.Equal(t, 7000, find(t, "switch.gnmi.sonic.port").Default)

	_, err = Update(b, map[string]any{"switch.gnmi.sonic.port": "8080", "switch.gnmi.sonic.plaintext": "false"}, nil, "tester")
	require.NoError(t, err)
	assert.Equal(t, 8080, SwitchGNMISonic.Port.Get())
	assert.False(t, SwitchGNMISonic.Plaintext.Get(), "an override can turn off what the shared setting turns on")
	assert.Equal(t, 7000, SwitchGNMIArista.Port.Get())
	assert.NotContains(t, Export(true), "[switch.gnmi.arista]", "a setting that only follows another is not written out")

	_, err = Update(b, nil, []string{"switch.gnmi.sonic.port"}, "tester")
	require.NoError(t, err)
	assert.Equal(t, 7000, SwitchGNMISonic.Port.Get())
}

func TestGeneratedSecretIsKeptInTheDatabase(t *testing.T) {
	b := &memBackend{rows: map[string]Row{}}
	start(t, b, nil, "")

	secret := APISecret.Get()
	assert.Len(t, secret, 64)
	assert.Len(t, ProvisionSecret.Get(), 32)
	assert.Contains(t, b.rows, "api.secret")

	got := find(t, "api.secret")
	assert.Equal(t, SourceDatabase, got.Source)
	assert.Nil(t, got.Value)
	assert.Nil(t, got.Stored)
	assert.Nil(t, got.Running)

	start(t, b, nil, "")
	assert.Equal(t, secret, APISecret.Get(), "a restart keeps the secret")

	start(t, b, map[string]any{"api.secret": "from-the-file"}, SourceFile)
	assert.Equal(t, "from-the-file", APISecret.Get())
}

func TestAnotherProcessChangeIsPickedUpOnReload(t *testing.T) {
	b := &memBackend{rows: map[string]Row{}}
	start(t, b, nil, "")
	changed := 0
	OnChange(func() { changed++ })

	b.rows["dns.forward"] = Row{Key: "dns.forward", Value: []byte(`"9.9.9.9:53"`)}
	b.rows["dns.ttl"] = Row{Key: "dns.ttl", Value: []byte(`600`)}
	b.revision++
	_, err := Reload(b)
	require.NoError(t, err)

	assert.Equal(t, 1, changed)
	assert.Equal(t, "9.9.9.9:53", DNSForward.Get())
	assert.Equal(t, 300, DNSTTL.Get(), "an option that needs a restart keeps what the process started with")
	assert.True(t, find(t, "dns.ttl").Pending)
}

func TestDerivedIsRebuiltOnlyWhenSettingsChange(t *testing.T) {
	b := &memBackend{rows: map[string]Row{}}
	start(t, b, nil, "")
	built := 0
	d := Derive(func() string { built++; return "forward=" + DNSForward.Get() })

	assert.Equal(t, "forward=", d.Get())
	assert.Equal(t, "forward=", d.Get())
	assert.Equal(t, 1, built)

	_, err := Update(b, map[string]any{"dns.forward": "1.1.1.1:53"}, nil, "tester")
	require.NoError(t, err)
	assert.Equal(t, "forward=1.1.1.1:53", d.Get())
	assert.Equal(t, 2, built)
}

func testFile(values map[string]any) File {
	f := File{Path: "old.toml", Lookup: func(key string) (any, bool) {
		v, ok := values[key]
		return v, ok
	}}
	for key := range values {
		f.Keys = append(f.Keys, key)
	}
	return f
}

func TestImport(t *testing.T) {
	b := &memBackend{rows: map[string]Row{}}
	start(t, b, nil, "")
	file := testFile(map[string]any{
		"dns.ttl":                   int64(900),
		"bmc.password":              "hunter2",
		"dbpath":                    "/var/lib/grendel/db",
		"dhcp.netmask":              int64(24),
		"provision.extra_vars":      map[string]any{"root_password": "s3cret", "timezone": "UTC"},
		"dns.tll":                   int64(5),
		"dns.listen":                "0.0.0.0:53",
		"bmc.switch_admin_username": "legacy",
		"client.api_endpoint":       "https://grendel.example",
		"dhcp.subnets":              []any{map[string]any{"gateway": "10.1.0.254/24"}, map[string]any{"gateway": "10.2.0.254/24"}},
	})

	plan, err := PlanImport(file, nil)
	require.NoError(t, err)
	actions := map[string]ImportItem{}
	for _, item := range plan.Items {
		actions[item.Key] = item
	}
	assert.Equal(t, ImportReady, actions["dns.ttl"].Action)
	assert.Equal(t, "900", actions["dns.ttl"].Value)
	assert.Equal(t, ImportSecret, actions["bmc.password"].Action)
	assert.Empty(t, actions["bmc.password"].Value)
	assert.Equal(t, "config file only", actions["bmc.password"].Note)
	assert.Equal(t, ImportExternal, actions["dbpath"].Action)
	assert.Equal(t, ImportClient, actions["client.api_endpoint"].Action)
	assert.Equal(t, ImportDeprecated, actions["dhcp.netmask"].Action)
	assert.Equal(t, "deprecated, it has no effect", actions["dhcp.netmask"].Note)
	assert.Equal(t, ImportExternal, actions["bmc.switch_admin_username"].Action)
	assert.Equal(t, "deprecated, kept in the config file as switch.username", actions["bmc.switch_admin_username"].Note)
	assert.Equal(t, ImportUnknown, actions["dns.tll"].Action)
	assert.Equal(t, "invalid setting", actions["dns.tll"].Note)
	assert.Equal(t, ImportDefault, actions["dns.listen"].Action)
	assert.Equal(t, "default value", actions["dns.listen"].Note)
	assert.Equal(t, ImportEntries, actions["provision.extra_vars"].Action)
	assert.Equal(t, "root_password, timezone", actions["provision.extra_vars"].Value, "entries are named but their values are not shown")
	assert.Equal(t, ImportReady, actions["dhcp.subnets"].Action, "subnets cannot be secrets, so they are imported like any other setting")
	assert.Equal(t, "10, 20", actions["dhcp.subnets"].Value)
	assert.Equal(t, 2, plan.Count())
	assert.Empty(t, plan.Invalid())

	remaining := plan.Remaining()
	assert.Contains(t, remaining, "dbpath = '/var/lib/grendel/db'")
	assert.Contains(t, remaining, "[bmc]\npassword = 'hunter2'")
	assert.Contains(t, remaining, "[client]\napi_endpoint = 'https://grendel.example'")
	assert.Contains(t, remaining, "[provision.extra_vars]\nroot_password = 's3cret'\ntimezone = 'UTC'")
	assert.Contains(t, remaining, "[switch]\nusername = 'legacy'", "a value under an old key is kept under the key that replaced it")
	for _, gone := range []string{"ttl", "tll", "listen", "netmask", "switch_admin_username", "subnets"} {
		assert.NotContains(t, remaining, gone)
	}

	plan, err = PlanImport(file, map[string][]string{"provision.extra_vars": {"timezone"}})
	require.NoError(t, err)
	assert.NotContains(t, plan.Remaining(), "timezone", "an imported entry leaves the file")
	assert.Contains(t, plan.Remaining(), "root_password")

	// the file being imported is usually the one in use, whose entries would otherwise be refused as the file's own
	start(t, b, map[string]any{"dhcp.subnets": []any{map[string]any{"gateway": "10.1.0.254/24"}}}, SourceFile)
	refused, err := plan.Apply(b, "tester")
	require.NoError(t, err)
	require.Empty(t, refused)
	assert.JSONEq(t, "900", string(b.rows["dns.ttl"].Value))
	assert.JSONEq(t, `{"timezone":"UTC"}`, string(b.rows["provision.extra_vars"].Value))
	assert.JSONEq(t, `"10.1.0.254/24"`, string(b.rows["dhcp.subnets.10.gateway"].Value))
	assert.JSONEq(t, `"10.2.0.254/24"`, string(b.rows["dhcp.subnets.20.gateway"].Value))
	assert.NotContains(t, b.rows, "switch.username")
	assert.NotContains(t, b.rows, "bmc.switch_admin_username")
	assert.NotContains(t, b.rows, "bmc.password")
	assert.NotContains(t, b.rows, "dns.listen")

	_, err = PlanImport(file, map[string][]string{"provision.extra_vars": {"missing"}})
	assert.ErrorContains(t, err, "no entry missing")
	_, err = PlanImport(file, map[string][]string{"dhcp.subnets": {"10"}})
	assert.Error(t, err)

	plan, err = PlanImport(testFile(map[string]any{"dns.query_timeout": "60d", "bmc.password": "hunter2"}), nil)
	require.NoError(t, err)
	require.Len(t, plan.Invalid(), 1)
	assert.Equal(t, "dns.query_timeout", plan.Invalid()[0].Key)
}

func TestExportWritesWhatTheDatabaseStores(t *testing.T) {
	b := &memBackend{rows: map[string]Row{}}
	start(t, b, nil, "")
	_, err := Update(b, map[string]any{
		"dns.ttl":                     "600",
		"dhcp.dns_servers":            `["10.0.0.1"]`,
		"dhcp.subnets.rack-f.gateway": "10.17.41.254/23",
		"dhcp.subnets.rack-f.mtu":     9000,
		"provision.extra_vars":        `{"timezone":"UTC"}`,
		"bmc.gofish.dial_timeout":     "7s",
	}, nil, "tester")
	require.NoError(t, err)

	stored := Export(false)
	assert.Contains(t, stored, "[dns]\nttl = 600\n")
	assert.Contains(t, stored, `dns_servers = ['10.0.0.1']`)
	assert.Contains(t, stored, "[dhcp.subnets.rack-f]\ngateway = '10.17.41.254/23'\nmtu = 9000")
	assert.Contains(t, stored, "[provision.extra_vars]\ntimezone = 'UTC'")
	assert.Contains(t, stored, "[bmc.gofish]\ndial_timeout = '7s'")
	assert.NotContains(t, stored, "secret")

	assert.NotContains(t, Export(true), "password")
}

func TestEveryOptionIsDescribed(t *testing.T) {
	for _, o := range All() {
		assert.NotEmpty(t, o.Summary, o.Key)
		assert.NotEmpty(t, o.Help, o.Key)
		_, err := o.Coerce(storedForm(o.Default))
		assert.NoError(t, err, "%s default", o.Key)
		if o.Kind == KindEnum {
			assert.NotEmpty(t, o.Enum, o.Key)
		}
		if o.Generated {
			assert.True(t, o.Secret, o.Key)
			assert.NotZero(t, o.GeneratedBytes, o.Key)
		}
	}
}

func TestNaturalCompare(t *testing.T) {
	names := []string{"rack-f", "10-core", "100", "2-lab", "10", "20", "15-lab", "rack-10", "rack-9"}
	slices.SortFunc(names, NaturalCompare)
	assert.Equal(t, []string{"2-lab", "10", "10-core", "15-lab", "20", "100", "rack-9", "rack-10", "rack-f"}, names)
}

func TestWrittenConfigParsesWhateverTheNamesAndValues(t *testing.T) {
	plan, err := PlanImport(testFile(map[string]any{
		"provision.extra_vars": map[string]any{"my key": "a", "motd": "line1\nbell\a \"quoted\"", "nested": map[string]any{"a.b": []any{"x"}}},
	}), nil)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, toml.Unmarshal([]byte(plan.Remaining()), &parsed), plan.Remaining())
	vars := parsed["provision"].(map[string]any)["extra_vars"].(map[string]any)
	assert.Equal(t, "a", vars["my key"])
	assert.Equal(t, "line1\nbell\a \"quoted\"", vars["motd"])
	assert.Equal(t, map[string]any{"a.b": []any{"x"}}, vars["nested"])
}

// racingBackend stands in for another process storing its own secrets between this one's first read and its insert
type racingBackend struct{ *memBackend }

func (r racingBackend) InsertSettings(set []Row, actor string) (int64, error) {
	for _, row := range set {
		r.rows[row.Key] = Row{Key: row.Key, Value: []byte(`"0123456789abcdef0123456789abcdef"`)}
	}
	return r.memBackend.InsertSettings(set, actor)
}

func TestGeneratedSecretIsTheOneTheDatabaseKept(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	_, err := Load(racingBackend{&memBackend{rows: map[string]Row{}}})
	require.NoError(t, err)
	assert.Equal(t, "0123456789abcdef0123456789abcdef", ProvisionSecret.Get(), "the secret another process stored first is used, not the one made up here")
}

func TestEntriesCannotComeFromTheEnvironment(t *testing.T) {
	problems := start(t, &memBackend{rows: map[string]Row{}}, map[string]any{"provision.extra_vars": `{"a":"b"}`}, SourceEnv)
	require.Len(t, problems, 1)
	assert.Equal(t, "provision.extra_vars", problems[0].Key)
	assert.Empty(t, ProvisionExtraVars.Get())
}

func TestRestoreLeavesOutWhatItCannotStore(t *testing.T) {
	b := &memBackend{rows: map[string]Row{}}
	start(t, b, nil, "")
	skipped, err := Restore(b, []Row{
		{Key: "dns.ttl", Value: []byte(`600`)},
		{Key: "dns.gone", Value: []byte(`1`)},
		{Key: "bmc.password", Value: []byte(`"x"`)},
		{Key: "dhcp.subnets.a.gateway", Value: []byte(`"10.1.0.254/24"`)},
		{Key: "dhcp.subnets.b.mtu", Value: []byte(`9000`)},
	}, "tester")
	require.NoError(t, err)
	assert.Equal(t, 600, find(t, "dns.ttl").Value)
	assert.Equal(t, map[string]any{"a": map[string]any{"gateway": "10.1.0.254/24"}}, DHCPSubnets.Get())
	assert.Len(t, b.userRows(), 2)
	assert.Contains(t, skipped, "dns.gone")
	assert.Contains(t, skipped, "bmc.password")
	assert.Contains(t, skipped, "dhcp.subnets.b.gateway", "an entry missing a field it needs is left out whole")
}

func TestBMCLoginsStayInTheConfigFile(t *testing.T) {
	b := &memBackend{rows: map[string]Row{}}
	start(t, b, nil, "")
	plan, err := PlanImport(testFile(map[string]any{
		"bmc.user":                 "root",
		"switch.snmp.user":         "monitor",
		"bmc.password":             "calvin",
		"bmc.credentials":          []any{map[string]any{"username": "Admin", "password": "x"}},
		"bmc.defaults.credentials": []any{map[string]any{"user": "ADMIN", "password": "ADMIN"}},
	}), nil)
	require.NoError(t, err)
	assert.Equal(t, 1, plan.Count(), "only the SNMP username can be stored")

	var kept map[string]any
	require.NoError(t, toml.Unmarshal([]byte(plan.Remaining()), &kept))
	bmc := kept["bmc"].(map[string]any)
	assert.Equal(t, "root", bmc["username"], "a login under the older key is kept under the one that replaced it")
	assert.NotContains(t, bmc, "user")
	assert.Equal(t, "calvin", bmc["password"])
	assert.Equal(t, []any{map[string]any{"username": "Admin", "password": "x"}}, bmc["credentials"], "a username keeps its case")
	assert.Contains(t, plan.Remaining(), "ADMIN", "logins under the older key are not dropped from the file")

	refused, err := plan.Apply(b, "tester")
	require.NoError(t, err)
	require.Empty(t, refused)
	assert.JSONEq(t, `"monitor"`, string(b.rows["switch.snmp.username"].Value), "a value under an old key is stored under the one that replaced it")

	refused, err = Update(b, map[string]any{"bmc.username": "ops", "switch.username": "ops"}, nil, "tester")
	require.NoError(t, err)
	assert.Contains(t, refused["bmc.username"], "outside the database")
	assert.Contains(t, refused["switch.username"], "outside the database")
}
