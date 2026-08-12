// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package migrations

import (
	"database/sql"

	"github.com/sirupsen/logrus"
)

// singleFQDNVersion is the migration that reduces nic.fqdn to a single name.
// Interfaces holding a comma separated list lose every name but the first.
const singleFQDNVersion = 20260811120000

// singleFQDNPreCheck reports the interface names that 20260811120000 is about
// to discard. The truncation itself lives in the migration's .sql file; this
// only reads and reports, and must run while the comma separated values are
// still intact.
var singleFQDNPreCheck = PreCheck{
	Version: singleFQDNVersion,
	Name:    "single-fqdn",
	Summary: "Interfaces now have a single FQDN. Only the first name was kept, the names logged above will no longer resolve",
	Report: func(db *sql.DB) ([]Finding, error) {
		rows, err := db.Query(`
			select coalesce(n.name, ''), nc.id, trim(substr(nc.fqdn, instr(nc.fqdn, ',') + 1))
			from nic as nc
			join node as n on n.id = nc.node_id
			where instr(coalesce(nc.fqdn, ''), ',') > 0
			order by nc.id`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		findings := make([]Finding, 0)
		for rows.Next() {
			var (
				nodeName string
				nicID    int64
				fqdn     string
			)
			if err := rows.Scan(&nodeName, &nicID, &fqdn); err != nil {
				return nil, err
			}

			findings = append(findings, Finding{
				Message: "Additional FQDN will be dropped.",
				Fields: logrus.Fields{
					"node":   nodeName,
					"nic_id": nicID,
					"fqdn":   fqdn,
				},
			})
		}

		return findings, rows.Err()
	},
}
