/*
 * SPDX-FileCopyrightText: (C) 2019 Grendel Authors
 *
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

-- name: ProvisionStatusFetch :one
select n.name, ps.state, ps.started_at, ps.updated_at, ps.completed_at, ps.detail
from provision_status as ps
join node as n on n.id = ps.node_id
where n.uid = @uid;

-- name: ProvisionStatusAll :many
select n.name, ps.state, ps.started_at, ps.updated_at, ps.completed_at, ps.detail
from provision_status as ps
join node as n on n.id = ps.node_id
order by n.name;

-- name: ProvisionStatusNodeset :many
select n.name, ps.state, ps.started_at, ps.updated_at, ps.completed_at, ps.detail
from provision_status as ps
join node as n on n.id = ps.node_id
where n.name in (sqlc.slice(nodeset))
order by n.name;

-- name: ProvisionStatusUpsert :execrows
insert into provision_status (node_id, state, started_at, updated_at, completed_at, detail)
select id, @state, @started_at, @updated_at, @completed_at, @detail
from node
where uid = @uid
on conflict (node_id)
do update set
  state = excluded.state,
  started_at = excluded.started_at,
  updated_at = excluded.updated_at,
  completed_at = excluded.completed_at,
  detail = excluded.detail;
