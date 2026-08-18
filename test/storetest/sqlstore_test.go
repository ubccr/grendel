// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package storetest

import (
	"database/sql"
	"path"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/sqlstore"
)

type SqlStoreTestSuite struct {
	StoreTestSuite
	file string
}

func TestSqlStoreTestSuite(t *testing.T) {
	store.Log.Logger.SetLevel(logrus.ErrorLevel)
	suite.Run(t, new(SqlStoreTestSuite))
}

func (s *SqlStoreTestSuite) SetFile(file string) {
	if file != ":memory:" {
		s.file = file
	}
}

func (s *SqlStoreTestSuite) SetupTest() {
	file := ":memory:"
	if s.file != "" {
		file = s.file
	}
	var err error
	ds, err := sqlstore.New(file)
	s.Assert().NoError(err)
	s.SetStore(ds)
}

func TestSqlStoreMigrations(t *testing.T) {
	assert := assert.New(t)

	file := path.Join(t.TempDir(), "grendel-test.db")

	var err error
	_, err = sqlstore.New(file)
	assert.NoError(err)
	_, err = sqlstore.New(file)
	assert.NoError(err)
	_, err = sqlstore.New(file)
	assert.NoError(err)
}

// TestReadOnlyDSNRejectsWrites guards the "file:" URI prefix in
// Config.DataSourceName.
//
// mattn/go-sqlite3 discards the entire query string unless the DSN starts with
// "file:" (sqlite3.go, in open(): `if !strings.HasPrefix(dsn, "file:") { dsn =
// dsn[:pos] }`), and it never reads "mode" itself. Without the prefix, mode=ro
// never reaches SQLite and the read pool silently opens read-write, because
// the driver always passes SQLITE_OPEN_READWRITE|SQLITE_OPEN_CREATE.
//
// Nothing in SqlStore writes through the read handle, so if this regressed no
// other test would notice. Raw SQL is deliberate: what is under test is what
// the connection permits, which the Store API cannot express.
func TestReadOnlyDSNRejectsWrites(t *testing.T) {
	file := path.Join(t.TempDir(), "grendel-test.db")
	cfg := sqlstore.Config{Driver: "sqlite3"}

	rw, err := sql.Open(cfg.Driver, cfg.DataSourceName(file, true))
	require.NoError(t, err)
	defer rw.Close()

	_, err = rw.Exec(`create table probe (x integer)`)
	require.NoError(t, err, "read-write handle must be writable")
	_, err = rw.Exec(`insert into probe values (1)`)
	require.NoError(t, err)

	ro, err := sql.Open(cfg.Driver, cfg.DataSourceName(file, false))
	require.NoError(t, err)
	defer ro.Close()

	var n int
	require.NoError(t, ro.QueryRow(`select count(*) from probe`).Scan(&n))
	assert.Equal(t, 1, n, "read-only handle must still serve reads")

	for _, stmt := range []string{
		`insert into probe values (2)`,
		`update probe set x = 9`,
		`delete from probe`,
		`create table sneaky (y integer)`,
		`drop table probe`,
		`pragma user_version = 7`,
	} {
		_, err := ro.Exec(stmt)
		assert.ErrorContains(t, err, "readonly database", "read-only handle accepted %q", stmt)
	}
}

// The "_"-prefixed parameters are mattn's own. It parses them out of the DSN
// before truncating the filename and applies them as pragmas after open, so
// they kept working even while mode=ro did not. Assert them so a later change
// to DataSourceName -- escaping the path, say -- cannot quietly drop WAL or
// foreign keys while still looking correct.
func TestDSNAppliesConnectionPragmas(t *testing.T) {
	file := path.Join(t.TempDir(), "grendel-test.db")
	cfg := sqlstore.Config{Driver: "sqlite3"}

	for _, rw := range []bool{true, false} {
		db, err := sql.Open(cfg.Driver, cfg.DataSourceName(file, rw))
		require.NoError(t, err)

		var journalMode string
		var foreignKeys int
		require.NoError(t, db.QueryRow(`pragma journal_mode`).Scan(&journalMode))
		require.NoError(t, db.QueryRow(`pragma foreign_keys`).Scan(&foreignKeys))

		assert.Equal(t, "wal", journalMode, "rw=%v", rw)
		assert.Equal(t, 1, foreignKeys, "rw=%v", rw)

		db.Close()
	}
}
