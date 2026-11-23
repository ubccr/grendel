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

func TestReplaceBmcRoutesKeepsRoleEdits(t *testing.T) {
	const before, after = 20261026120000, 20261110120000
	db, m := migratedTo(t, before)

	// an admin took power and clearing the event log away from the user role, and made a role that only reads BMCs
	revoke(t, db, "user", "POST", "/v1/bmc/power/os")
	revoke(t, db, "user", "POST", "/v1/bmc/power/bmc")
	revoke(t, db, "user", "DELETE", "/v1/bmc/sel")
	exec(t, db, `insert into role(name) values ('ops')`)
	grant(t, db, "ops", "GET", "/v1/bmc")

	move(t, m, after)

	assertGrants(t, db, "admin", "/v1/bmc",
		"GET /v1/bmc/status", "GET /v1/bmc/sel", "GET /v1/bmc/firmware/updates", "GET /v1/bmc/jobs",
		"POST /v1/bmc/power", "POST /v1/bmc/indicator", "POST /v1/bmc/jobs/clear", "POST /v1/bmc/sel/clear", "POST /v1/bmc/firmware/install",
		"POST /v1/bmc/configure/auto", "POST /v1/bmc/configure/import")
	assertGrants(t, db, "user", "/v1/bmc",
		"GET /v1/bmc/status", "GET /v1/bmc/sel", "GET /v1/bmc/firmware/updates", "GET /v1/bmc/jobs",
		"POST /v1/bmc/jobs/clear", "POST /v1/bmc/firmware/install",
		"POST /v1/bmc/configure/auto", "POST /v1/bmc/configure/import")
	assertGrants(t, db, "read-only", "/v1/bmc", "GET /v1/bmc/status", "GET /v1/bmc/sel", "GET /v1/bmc/firmware/updates", "GET /v1/bmc/jobs")
	assertGrants(t, db, "ops", "/v1/bmc", "GET /v1/bmc/status", "GET /v1/bmc/sel", "GET /v1/bmc/firmware/updates")

	move(t, m, before)

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

func TestGrantUserTagsDownKeepsEarlierGrants(t *testing.T) {
	const before, after = 20261117120000, 20261118120000
	db, m := migratedTo(t, before)

	// an admin had already let users list roles
	grant(t, db, "user", "GET", "/v1/roles")

	move(t, m, after)
	assertGrants(t, db, "user", "/v1/nodes/tags/", "PATCH /v1/nodes/tags/%")
	assertGrants(t, db, "user", "/v1/roles", "GET /v1/roles")

	move(t, m, before)
	assertGrants(t, db, "user", "/v1/roles", "GET /v1/roles")
}

func TestCountEventsFromNodesRoundTrip(t *testing.T) {
	const before, after = 20261121120000, 20261122120000
	db, m := migratedTo(t, before)

	exec(t, db, `insert into event(id, kind, action, summary, actor, state, severity, started_at, updated_at, counts) values
		('e1', 'bmc', 'bmc.power', 'bmc power', 'admin', 'done', 'warning', '2026-11-16 12:00:00', '2026-11-16 12:00:00', '{"total":2,"pending":0,"running":0,"succeeded":1,"failed":1,"stalled":0}'),
		('e2', 'audit', 'node.save', 'Saved nodes', 'admin', 'done', 'success', '2026-11-16 12:00:00', '2026-11-16 12:00:00', '{}')`)
	exec(t, db, `insert into event_node(event_id, node, state, updated_at) values
		('e1', 'n1', 'succeeded', '2026-11-16 12:00:00'),
		('e1', 'n2', 'failed', '2026-11-16 12:00:00')`)

	counts := func(id string) string {
		t.Helper()
		var c string
		if err := db.QueryRow(`select counts from event where id = ?`, id).Scan(&c); err != nil {
			t.Fatal(err)
		}
		return c
	}
	want := counts("e1")

	move(t, m, after)
	if _, err := db.Exec(`select counts from event`); err == nil {
		t.Error("the counts column is still there after the up")
	}
	var open int
	if err := db.QueryRow(`select count(*) from sqlite_master where type = 'index' and name = 'event_node_open'`).Scan(&open); err != nil || open != 1 {
		t.Errorf("event_node_open index: %d, %v", open, err)
	}

	move(t, m, before)
	if got := counts("e1"); got != want {
		t.Errorf("counts after the round trip are %s, want %s", got, want)
	}
	if got := counts("e2"); got != "{}" {
		t.Errorf("an event with no nodes has counts %s, want {}", got)
	}
	if err := db.QueryRow(`select count(*) from sqlite_master where type = 'index' and name = 'event_node_open'`).Scan(&open); err != nil || open != 0 {
		t.Errorf("event_node_open index after the down: %d, %v", open, err)
	}
}

func TestReassignDefaultPermissionsRoundTrip(t *testing.T) {
	const before, after = 20261122120000, 20261123120000
	db, m := migratedTo(t, before)
	readOnlySwitch := grants(t, db, "read-only", "/v1/switch")
	userUsers := grants(t, db, "user", "/v1/users")

	move(t, m, after)

	assertGrants(t, db, "read-only", "/v1/switch", "GET /v1/switch/%/lldp", "GET /v1/switch/scan")
	assertGrants(t, db, "read-only", "/v1/nodes/ips", "GET /v1/nodes/ips/next", "GET /v1/nodes/ips/ranges")
	assertGrants(t, db, "read-only", "/v1/roles", "GET /v1/roles")
	assertGrants(t, db, "user", "/v1/users")

	move(t, m, before)

	assertGrants(t, db, "read-only", "/v1/switch", readOnlySwitch...)
	assertGrants(t, db, "read-only", "/v1/nodes/ips", "GET /v1/nodes/ips/ranges")
	assertGrants(t, db, "read-only", "/v1/roles")
	assertGrants(t, db, "user", "/v1/users", userUsers...)
}
