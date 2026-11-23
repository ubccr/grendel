/*
 * SPDX-FileCopyrightText: (C) 2019 Grendel Authors
 *
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

-- name: EventUpsert :exec
insert into event (id, kind, action, summary, subject, actor, state, severity, started_at, updated_at, ended_at, detail)
values (@id, @kind, @action, @summary, @subject, @actor, @state, @severity, @started_at, @updated_at, @ended_at, @detail)
on conflict (id)
do update set
  summary = excluded.summary,
  subject = excluded.subject,
  state = excluded.state,
  severity = excluded.severity,
  updated_at = excluded.updated_at,
  ended_at = excluded.ended_at,
  detail = excluded.detail;

-- name: EventFetch :one
select * from event where id = @id;

-- name: EventTouchOpen :one
-- says the newest running event of a kind was updated, and returns its id
update event set updated_at = @updated_at
where id = (
  select o.id from event o
  where o.kind = @kind and o.state = 'running'
  order by o.started_at desc, o.id desc
  limit 1
)
returning id;

-- name: EventList :many
select * from event
where (cast(@state as text) = '' or state = cast(@state as text))
  and (cast(@kind as text) = '' or kind = cast(@kind as text))
  and (cast(@before as text) = '' or (started_at, id) < (select b.started_at, b.id from event b where b.id = cast(@before as text)))
-- ids only order events to the second, so they are ordered by when they started and the id only breaks ties
order by started_at desc, id desc
limit @max;

-- name: EventNodeUpsert :exec
insert into event_node (event_id, node, state, stage, message, ref, updated_at)
values (@event_id, @node, @state, @stage, @message, @ref, @updated_at)
on conflict (event_id, node)
do update set
  state = excluded.state,
  stage = excluded.stage,
  message = excluded.message,
  ref = excluded.ref,
  updated_at = excluded.updated_at;

-- name: EventNodeLatest :one
select e.id
from event_node n
join event e on e.id = n.event_id
where e.kind = @kind and n.node = @node
order by e.started_at desc, e.id desc
limit 1;

-- name: EventNodes :many
select * from event_node where event_id = @event_id;

-- name: EventNodeCounts :many
select event_id, state, stage, count(*) as nodes
from event_node
where event_id in (sqlc.slice(ids))
group by event_id, state, stage;

-- name: EventNodeOpen :one
select exists (
  select 1 from event_node
  where event_id = @event_id and state in ('pending', 'running')
);

-- name: EventNodesStall :execrows
update event_node
set state = 'stalled'
where event_node.state in ('pending', 'running')
  and event_node.updated_at < @before
  and event_id in (select id from event where kind = @kind and event.state = 'running');

-- name: EventPrune :execrows
delete from event where state = 'done' and started_at < @before;

-- name: EventTouch :execrows
update event set updated_at = @updated_at where id = @id;

-- name: EventNodesIn :many
select * from event_node where event_id in (sqlc.slice(ids));
