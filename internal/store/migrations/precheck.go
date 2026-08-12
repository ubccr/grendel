// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package migrations

import (
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/internal/logger"
)

var log = logger.GetLogger("MIGRATE")

// A PreCheck inspects the database immediately before a migration is applied,
// while the old schema is still intact, and reports what that migration is
// about to change. It exists for destructive migrations, where the log it
// produces is the only record an operator gets of what was discarded.
//
// Checks are advisory. They never mutate the database and never block a
// migration — the .sql files remain the sole source of truth for what actually
// changes. A check that fails is logged and skipped, because a database old
// enough to be missing the tables a check queries is a normal thing to
// encounter, not an error.
type PreCheck struct {
	// Version is the migration this check belongs to. The check runs only when the database is below this version, i.e. when that migration is pending, so it costs nothing on subsequent startups.
	Version uint

	// Name identifies the check in logs.
	Name string

	// Summary is logged once, with the finding count, when Report returns any findings. It should tell the operator what the consequence is.
	Summary string

	// Report returns one finding per affected row. Returning an error means the check could not run, not that the database is invalid.
	Report func(db *sql.DB) ([]Finding, error)
}

// A Finding is one thing a pending migration will change. Fields are logged as structured keys so they can be grepped or shipped to a log aggregator.
type Finding struct {
	Message string
	Fields  logrus.Fields
}

// preChecks is the registry. Add new checks here; ordering follows Version.
var preChecks = []PreCheck{
	singleFQDNPreCheck,
}

// runPreChecks reports on every registered check whose migration is still pending. Called by Migrate before anything is applied.
func runPreChecks(db *sql.DB, cur uint) {
	for _, preCheck := range preChecks {
		if cur >= preCheck.Version {
			continue
		}

		findings, err := preCheck.Report(db)
		if err != nil {
			// Most often this is a database predating the tables the check queries, which is expected and not worth alarming about.
			log.WithFields(logrus.Fields{
				"check": preCheck.Name,
				"err":   err,
			}).Debug("Skipped migration pre-check")
			continue
		}
		if len(findings) == 0 {
			continue
		}

		for _, f := range findings {
			log.WithFields(f.Fields).Warn(f.Message)
		}
		log.WithFields(logrus.Fields{
			"check":    preCheck.Name,
			"findings": len(findings),
			"version":  preCheck.Version,
		}).Warn(preCheck.Summary)
	}
}
