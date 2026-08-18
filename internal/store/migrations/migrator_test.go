// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package migrations

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

// TestMigratePreCheckFiresBeforeTruncation stages a database at the version
// immediately before the single-FQDN migration, with a multi-name interface in
// it, and migrates to exactly that version. It proves the two halves work
// together: the Go side reports the names, the .sql file discards them.
func TestMigratePreCheckFiresBeforeTruncation(t *testing.T) {
	file := filepath.Join(t.TempDir(), "grendel.db")
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Just enough of the old schema for 20260811120000 to apply, pinned to the
	// version immediately before it so only that one migration runs. mac is
	// here only because the migration indexes it.
	if _, err := db.Exec(`
		create table schema_migrations (version uint64, dirty bool);
		insert into schema_migrations values (20260519210025, false);

		create table node (id integer primary key, name text);
		create table nic (
			id integer primary key,
			node_id integer not null,
			fqdn text,
			ip text,
			mac text
		);
		insert into node(id, name) values (1, 'cpn-42');
		insert into nic(id, node_id, fqdn, ip) values
		  (1, 1, 'cpn-42.cluster.local,cpn-42-ib.cluster.local', '10.1.1.17/24'),
		  (2, 1, '', '10.1.1.18/24');

		-- The migration drops and recreates node_view. A stub stands in for it
		-- here; the recreated view references tables this staged schema does not
		-- have, which is fine because SQLite does not validate a view until it
		-- is queried, and this test never queries it.
		create view node_view as select 1 as id;`); err != nil {
		t.Fatal(err)
	}

	hook := logrustest.NewLocal(log.Logger)
	defer hook.Reset()

	m, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	// the staged schema above only carries what 20260811120000 touches, so this stops there rather than following SchemaVersion into migrations that need tables this test never creates
	if err := m.migrateTo(20260811120000); err != nil && err != ErrNoChange {
		t.Fatal(err)
	}

	t.Run("reported the dropped name before discarding it", func(t *testing.T) {
		var found *logrus.Entry
		for _, e := range hook.AllEntries() {
			if e.Data["fqdn"] == "cpn-42-ib.cluster.local" {
				found = e
				break
			}
		}
		if found == nil {
			t.Fatalf("no warning naming the dropped alias; got %d entries", len(hook.AllEntries()))
		}
		if found.Level != logrus.WarnLevel {
			t.Errorf("logged at %s, want warn", found.Level)
		}
		if found.Data["node"] != "cpn-42" || found.Data["nic_id"] != int64(1) {
			t.Errorf("missing node/nic context: %+v", found.Data)
		}
	})

	t.Run("logged a summary", func(t *testing.T) {
		for _, e := range hook.AllEntries() {
			if strings.Contains(e.Message, "single FQDN") {
				return
			}
		}
		t.Error("no summary warning explaining the consequence")
	})

	t.Run("the sql truncated the value", func(t *testing.T) {
		var fqdn sql.NullString
		if err := db.QueryRow(`select fqdn from nic where id = 1`).Scan(&fqdn); err != nil {
			t.Fatal(err)
		}
		if fqdn.String != "cpn-42.cluster.local" {
			t.Errorf("got %q, want %q", fqdn.String, "cpn-42.cluster.local")
		}
	})

	t.Run("empty names became null", func(t *testing.T) {
		var fqdn sql.NullString
		if err := db.QueryRow(`select fqdn from nic where id = 2`).Scan(&fqdn); err != nil {
			t.Fatal(err)
		}
		if fqdn.Valid {
			t.Errorf("got %q, want NULL", fqdn.String)
		}
	})

	t.Run("the CIDR was split into address and prefix", func(t *testing.T) {
		var (
			addr   string
			prefix int
		)
		if err := db.QueryRow(`select ip, prefix_len from nic where id = 1`).Scan(&addr, &prefix); err != nil {
			t.Fatal(err)
		}
		if addr != "10.1.1.17" || prefix != 24 {
			t.Errorf("got %q/%d, want 10.1.1.17/24", addr, prefix)
		}
	})

}

// A database already at the target version must not re-run the check.
func TestMigratePreCheckSkippedWhenUpToDate(t *testing.T) {
	file := filepath.Join(t.TempDir(), "grendel.db")
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		create table schema_migrations (version uint64, dirty bool);
		insert into schema_migrations values (` + strconv.FormatUint(SchemaVersion, 10) + `, false);
		create table node (id integer primary key, name text);
		create table nic (id integer primary key, node_id integer not null, fqdn text);
		insert into node(id, name) values (1, 'cpn-42');
		insert into nic(id, node_id, fqdn) values (1, 1, 'a.example.com,b.example.com');`); err != nil {
		t.Fatal(err)
	}

	hook := logrustest.NewLocal(log.Logger)
	defer hook.Reset()

	m, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	if err := m.Migrate(); err != nil && err != ErrNoChange {
		t.Fatal(err)
	}

	for _, e := range hook.AllEntries() {
		if e.Data["dropped"] != nil {
			t.Errorf("pre-check ran against an already-migrated database: %+v", e.Data)
		}
	}
}
