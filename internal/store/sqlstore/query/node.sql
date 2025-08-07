/*
 * SPDX-FileCopyrightText: (C) 2019 Grendel Authors
 *
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

-- name: NodeCount :one
select count(*) from node;

-- name: NodeFetchByID :one
select * from node_view where id = @id;

-- name: NodeFetchByUID :one
select * from node_view where uid = @uid;

-- name: NodeFetchByName :one
select * from node_view where name = @name;

-- name: NodeFind :one
select n.* 
from node_view as n
join nic as nc
on nc.node_id = n.id
where 
  case 
    when cast(@filter_mac as integer) then nc.mac = @mac
    when cast(@filter_ip as integer) then nc.ip = @ip
    else 0
  end
limit 1;

-- name: NodeResolve :many
select nc.fqdn, nc.ip 
from nic as nc
where 
  case 
    when cast(@filter_fqdn as integer) then lower(nc.fqdn) like concat('%', cast(@fqdn as text), '%')
    when cast(@filter_ip as integer) then substring(nc.ip, 0, instr(nc.ip, '/')) = cast(@ip as text)
    else 0
  end;

-- name: NodeAll :many
select * from node_view;

-- name: NodeID :many
select id from node
where name in (sqlc.slice(nodeset));

-- name: TagID :many
select id from tag
where key in (sqlc.slice(tags));

-- name: NodeFindNodeset :many
select * from node_view 
where name in (sqlc.slice(nodeset));

-- name: NodeFindTags :many
select 
  n.name as name,
  count(distinct(t.key)) as cnt
from
  node as n
join node_tag as nt
  on nt.node_id = n.id 
join tag as t
  on nt.tag_id = t.id 
where 
  t.key in (sqlc.slice(tags))
group by name;

-- name: NodeProvision :exec
update node set provision = @provision
where id in (sqlc.slice(nodes));

-- name: NodeBootKernel :exec
update node set kernel_id = @kernel_id
where id in (sqlc.slice(nodes));

-- name: NodeUpsert :one
insert into node (id, uid, name, provision, arch_id, kernel_id, node_type_id, firmware)
values (sqlc.narg('id'), @uid, @name, @provision, @arch_id, @kernel_id, @node_type_id, @firmware)
on conflict (id)
do update set name = ?3, provision = ?4, arch_id = ?5, kernel_id = ?6, node_type_id = ?7, firmware = ?8
returning *;

-- name: NodeDelete :exec
delete from node where name in (sqlc.slice(nodeset));

-- name: TagUpsert :one
insert into tag (key)
values (@key)
on conflict (key)
do update set key = ?1
returning *;

-- name: NodeTagUpsert :exec
insert into node_tag (tag_id, node_id, value)
values (@tag_id, @node_id, @value)
on conflict (tag_id, node_id, value)
do nothing;

-- name: NodeTagUpsertDelete :exec
delete from node_tag where node_id in (sqlc.slice(nodes)) and tag_id not in (sqlc.slice(tags));

-- name: NodeTagDelete :exec
delete from node_tag where node_id in (sqlc.slice(nodes)) and tag_id in (sqlc.slice(tags));
