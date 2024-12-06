// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed sql/*.sql
var sqlMigrations embed.FS

type Migrator struct {
	mg     *migrate.Migrate
	source source.Driver
}

func New(db *sql.DB) (*Migrator, error) {
	source, err := iofs.New(sqlMigrations, "sql")
	if err != nil {
		return nil, fmt.Errorf("Failed to find sql migration sources: %w", err)
	}

	dbi, err := sqlite.WithInstance(db, new(sqlite.Config))
	if err != nil {
		return nil, fmt.Errorf("Invalid sqlite db instance: %w", err)
	}

	mg, err := migrate.NewWithInstance("iofs", source, "sqlite", dbi)
	if err != nil {
		return nil, fmt.Errorf("Failed to create migrate instance: %w", err)
	}

	return &Migrator{mg: mg, source: source}, nil
}

func (m *Migrator) Version() (uint, bool, error) {
	ver, dirty, err := m.mg.Version()
	if err != nil && err == migrate.ErrNilVersion {
		return ver, dirty, ErrNilVersion
	}

	return ver, dirty, err
}

func (m *Migrator) Migrate() error {
	err := m.mg.Migrate(SchemaVersion)
	if err != nil && err == migrate.ErrNoChange {
		return ErrNoChange
	}

	return err
}

func (m *Migrator) Close() error {
	return m.source.Close()
}
