// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: nic.sql

package db

import (
	"context"

	null "github.com/guregu/null/v5"
)

const nicUpsert = `-- name: NicUpsert :one
/*
 * SPDX-FileCopyrightText: (C) 2019 Grendel Authors
 *
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

insert into nic (id, node_id, nic_type, name, vlan, fqdn, mac, ip, peers, mtu)
values (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10)
on conflict (id)
do update set nic_type = ?3, name = ?4, vlan = ?5, fqdn = ?6, mac = ?7, ip = ?8,
              peers = ?9, mtu = ?10
returning id, node_id, nic_type, name, vlan, fqdn, mac, ip, peers, mtu
`

type NicUpsertParams struct {
	ID      null.Int64  `json:"id"`
	NodeID  int64       `json:"node_id"`
	NicType string      `json:"nic_type"`
	Name    null.String `json:"name"`
	VLAN    null.String `json:"vlan"`
	FQDN    null.String `json:"fqdn"`
	MAC     null.String `json:"mac"`
	IP      null.String `json:"ip"`
	Peers   null.String `json:"peers"`
	MTU     null.Int64  `json:"mtu"`
}

func (q *Queries) NicUpsert(ctx context.Context, db DBTX, arg NicUpsertParams) (Nic, error) {
	row := db.QueryRowContext(ctx, nicUpsert,
		arg.ID,
		arg.NodeID,
		arg.NicType,
		arg.Name,
		arg.VLAN,
		arg.FQDN,
		arg.MAC,
		arg.IP,
		arg.Peers,
		arg.MTU,
	)
	var i Nic
	err := row.Scan(
		&i.ID,
		&i.NodeID,
		&i.NicType,
		&i.Name,
		&i.VLAN,
		&i.FQDN,
		&i.MAC,
		&i.IP,
		&i.Peers,
		&i.MTU,
	)
	return i, err
}