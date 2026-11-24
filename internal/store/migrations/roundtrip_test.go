// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package migrations

import (
	"database/sql"
	"path/filepath"
	"slices"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// migratedTo opens a fresh database migrated up to version
func migratedTo(t *testing.T, version uint) (*sql.DB, *Migrator) {
	t.Helper()
	db, err := sql.Open("sqlite3", "file:"+filepath.Join(t.TempDir(), "grendel.db")+"?_foreign_keys=true")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	m, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })
	move(t, m, version)
	return db, m
}

func move(t *testing.T, m *Migrator, version uint) {
	t.Helper()
	if err := m.migrateTo(version); err != nil && err != ErrNoChange {
		t.Fatalf("migrating to %d: %s", version, err)
	}
}

func exec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("%s: %s", query, err)
	}
}

// grants lists what a role may do under a path prefix, as "METHOD path", sorted
func grants(t *testing.T, db *sql.DB, role, prefix string) []string {
	t.Helper()
	rows, err := db.Query(`
		select p.method || ' ' || p.path
		from role_permission rp
		join role r on r.id = rp.role_id
		join permission p on p.id = rp.permission_id
		where r.name = ? and p.path like ? || '%'
		order by 1`, role, prefix)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			t.Fatal(err)
		}
		out = append(out, g)
	}
	return out
}

func revoke(t *testing.T, db *sql.DB, role, method, path string) {
	t.Helper()
	exec(t, db, `delete from role_permission
		where role_id = (select id from role where name = ?)
		  and permission_id = (select id from permission where (method, path) = (?, ?))`, role, method, path)
}

func grant(t *testing.T, db *sql.DB, role, method, path string) {
	t.Helper()
	exec(t, db, `insert into role_permission(role_id, permission_id)
		select role.id, permission.id from role, permission
		where role.name = ? and (permission.method, permission.path) = (?, ?)`, role, method, path)
}

func assertGrants(t *testing.T, db *sql.DB, role, prefix string, want ...string) {
	t.Helper()
	slices.Sort(want)
	if got := grants(t, db, role, prefix); !slices.Equal(got, want) {
		t.Errorf("%s under %s holds\n  %q\nwant\n  %q", role, prefix, got, want)
	}
}

// the schema v0.2.10 leaves and the one 0.3.0 migrates it to
const before030, after030 = 20261026120000, 20261126120000

func TestRelease030KeepsRoleEdits(t *testing.T) {
	db, m := migratedTo(t, before030)

	// an admin took power and clearing the event log away from the user role, and made a role that only reads BMCs
	revoke(t, db, "user", "POST", "/v1/bmc/power/os")
	revoke(t, db, "user", "POST", "/v1/bmc/power/bmc")
	revoke(t, db, "user", "DELETE", "/v1/bmc/sel")
	exec(t, db, `insert into role(name) values ('ops')`)
	grant(t, db, "ops", "GET", "/v1/bmc")

	move(t, m, after030)

	reads := []string{"GET /v1/bmc/status", "GET /v1/bmc/sel", "GET /v1/bmc/firmware/updates", "GET /v1/bmc/nics", "GET /v1/bmc/jobs", "GET /v1/bmc/bios", "GET /v1/bmc/bios/registry", "GET /v1/bmc/identity/defaults", "GET /v1/bmc/secureboot"}
	userWrites := []string{"POST /v1/bmc/jobs/clear", "POST /v1/bmc/firmware/install", "POST /v1/bmc/configure/auto", "POST /v1/bmc/configure/import", "POST /v1/bmc/bios", "POST /v1/bmc/bios/discard", "POST /v1/bmc/asset-tag", "POST /v1/bmc/location", "GET /v1/bmc/console"}
	adminOnly := []string{"POST /v1/bmc/power", "POST /v1/bmc/indicator", "POST /v1/bmc/sel/clear", "POST /v1/bmc/secureboot", "POST /v1/bmc/secureboot/keys", "POST /v1/bmc/secureboot/keys/delete", "POST /v1/bmc/secureboot/reset"}
	assertGrants(t, db, "admin", "/v1/bmc", slices.Concat(reads, userWrites, adminOnly)...)
	assertGrants(t, db, "user", "/v1/bmc", slices.Concat(reads, userWrites)...)
	assertGrants(t, db, "read-only", "/v1/bmc", reads...)
	assertGrants(t, db, "ops", "/v1/bmc", "GET /v1/bmc/status", "GET /v1/bmc/sel", "GET /v1/bmc/firmware/updates", "GET /v1/bmc/nics")

	move(t, m, before030)

	assertGrants(t, db, "user", "/v1/bmc",
		"GET /v1/bmc", "GET /v1/bmc/jobs",
		"DELETE /v1/bmc/jobs", "DELETE /v1/bmc/jobs/%",
		"GET /v1/bmc/upgrade/dell/repo", "POST /v1/bmc/upgrade/dell/installfromrepo",
		"POST /v1/bmc/configure/auto", "POST /v1/bmc/configure/import")
	assertGrants(t, db, "read-only", "/v1/bmc", "GET /v1/bmc", "GET /v1/bmc/jobs")
	assertGrants(t, db, "ops", "/v1/bmc", "GET /v1/bmc")
	if got := grants(t, db, "admin", "/v1/bmc/metrics"); len(got) != 1 {
		t.Errorf("admin holds %q of metrics, want it back", got)
	}
}

func TestRelease030DefaultGrantsRoundTrip(t *testing.T) {
	db, m := migratedTo(t, before030)
	readOnlySwitch := grants(t, db, "read-only", "/v1/switch")
	userUsers := grants(t, db, "user", "/v1/users")
	adminSwitch := grants(t, db, "admin", "/v1/switch")

	move(t, m, after030)

	var total, admin int
	if err := db.QueryRow(`select (select count(*) from permission), (select count(*) from role_permission rp join role r on r.id = rp.role_id where r.name = 'admin')`).Scan(&total, &admin); err != nil || admin != total {
		t.Errorf("admin holds %d of %d permissions, %v", admin, total, err)
	}
	assertGrants(t, db, "read-only", "/v1/switch", "GET /v1/switch/%/lldp", "GET /v1/switch/scan")
	assertGrants(t, db, "read-only", "/v1/nodes/ips", "GET /v1/nodes/ips/next", "GET /v1/nodes/ips/ranges")
	assertGrants(t, db, "read-only", "/v1/roles", "GET /v1/roles")
	assertGrants(t, db, "user", "/v1/nodes/tags", "GET /v1/nodes/tags", "PATCH /v1/nodes/tags/%")
	assertGrants(t, db, "user", "/v1/users")
	assertGrants(t, db, "user", "/v1/nodes/render")
	assertGrants(t, db, "user", "/v1/settings")
	assertGrants(t, db, "read-only", "/v1/settings")
	assertGrants(t, db, "admin", "/v1/settings", "GET /v1/settings", "PATCH /v1/settings")

	move(t, m, before030)

	assertGrants(t, db, "admin", "/v1/switch", adminSwitch...)
	assertGrants(t, db, "read-only", "/v1/switch", readOnlySwitch...)
	assertGrants(t, db, "read-only", "/v1/roles")
	assertGrants(t, db, "user", "/v1/users", userUsers...)
	var left int
	if err := db.QueryRow(`select count(*) from permission where path like '/v1/settings%' or path like '/v1/nodes/ips/%' or path like '/v1/bmc/bios%'`).Scan(&left); err != nil || left != 0 {
		t.Errorf("permissions of 0.3.0 left after the down: %d, %v", left, err)
	}
}

func TestRelease030DownKeepsEarlierUserGrants(t *testing.T) {
	db, m := migratedTo(t, before030)

	// an admin had already let users list roles
	grant(t, db, "user", "GET", "/v1/roles")

	move(t, m, after030)
	assertGrants(t, db, "user", "/v1/roles", "GET /v1/roles")

	move(t, m, before030)
	assertGrants(t, db, "user", "/v1/roles", "GET /v1/roles")
}

func TestRelease030TablesRoundTrip(t *testing.T) {
	db, m := migratedTo(t, before030)
	tables := func() []string {
		t.Helper()
		rows, err := db.Query(`select name from sqlite_master where type in ('table', 'index') and name in ('provision_status', 'event', 'event_node', 'event_node_open', 'event_open_provision', 'setting', 'setting_revision') order by 1`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		out := []string{}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			out = append(out, name)
		}
		return out
	}

	move(t, m, after030)
	if got, want := tables(), []string{"event", "event_node", "event_node_open", "event_open_provision", "provision_status", "setting", "setting_revision"}; !slices.Equal(got, want) {
		t.Errorf("after the up there is %q, want %q", got, want)
	}
	var revision int
	if err := db.QueryRow(`select revision from setting_revision`).Scan(&revision); err != nil || revision != 0 {
		t.Errorf("setting revision starts at %d, %v", revision, err)
	}

	// rows in the new tables do not stop the down
	exec(t, db, `insert into event(id, kind, action, summary, actor, state, severity, started_at, updated_at) values ('e1', 'bmc', 'bmc.power', 'bmc power', 'admin', 'done', 'warning', '2026-11-16 12:00:00', '2026-11-16 12:00:00')`)
	exec(t, db, `insert into event_node(event_id, node, state, updated_at) values ('e1', 'n1', 'succeeded', '2026-11-16 12:00:00')`)
	exec(t, db, `insert into setting(key, value) values ('dns.ttl', '600')`)

	move(t, m, before030)
	if got := tables(); len(got) != 0 {
		t.Errorf("after the down there is still %q", got)
	}
}
