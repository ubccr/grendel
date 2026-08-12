// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package migrations

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// openPreMigrationDB builds the shape of the database as it exists just before
// 20260811120000: nic.fqdn is a plain column that may hold a comma separated
// list of names.
func openPreMigrationDB(t *testing.T) *sql.DB {
	t.Helper()

	file := filepath.Join(t.TempDir(), "grendel.db")
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSingleFQDNPreCheck(t *testing.T) {
	db := openPreMigrationDB(t)

	if _, err := db.Exec(`
		create table node (id integer primary key, name text);
		create table nic (id integer primary key, node_id integer not null, fqdn text);
		insert into node(id, name) values (1, 'cpn-42'), (2, 'cpn-43');
		insert into nic(id, node_id, fqdn) values
		  (1, 1, 'cpn-42.cluster.local, cpn-42-ib.cluster.local , cpn-42-eth.cluster.local'),
		  (2, 1, 'bmc-42.cluster.local'),
		  (3, 2, null),
		  (4, 2, '');`); err != nil {
		t.Fatal(err)
	}

	findings, err := singleFQDNPreCheck.Report(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 multi-name interface, got %d: %+v", len(findings), findings)
	}

	got := findings[0]
	if got.Fields["node"] != "cpn-42" || got.Fields["nic_id"] != int64(1) {
		t.Errorf("expected node cpn-42 nic 1, got %+v", got.Fields)
	}

	// Everything after the first comma, which is exactly what the migration
	// throws away.
	want := "cpn-42-ib.cluster.local , cpn-42-eth.cluster.local"
	if got.Fields["fqdn"] != want {
		t.Errorf("dropped names:\n got %q\nwant %q", got.Fields["fqdn"], want)
	}
}

// A database old enough to predate the nic table has nothing to report. That is
// an expected state, not an error, and it must not stop the migration.
func TestSingleFQDNPreCheckNoNicTable(t *testing.T) {
	db := openPreMigrationDB(t)

	if _, err := singleFQDNPreCheck.Report(db); err == nil {
		t.Fatal("expected an error querying a database with no nic table")
	}

	// runPreChecks swallows it.
	runPreChecks(db, 0)
}

// Checks run only while their migration is pending.
func TestRunPreChecksSkipsAppliedMigrations(t *testing.T) {
	db := openPreMigrationDB(t)

	ran := false
	original := preChecks
	preChecks = []PreCheck{{
		Version: 100,
		Name:    "test",
		Summary: "test",
		Report: func(*sql.DB) ([]Finding, error) {
			ran = true
			return nil, nil
		},
	}}
	t.Cleanup(func() { preChecks = original })

	runPreChecks(db, 100)
	if ran {
		t.Error("check ran even though its migration was already applied")
	}

	runPreChecks(db, 99)
	if !ran {
		t.Error("check did not run for a pending migration")
	}
}
