-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

-- only a node's latest provision is kept, a new network boot overwrites the row
create table provision_status (
  node_id       integer primary key,
  state         text      not null,
  started_at    timestamp not null,
  updated_at    timestamp not null,
  completed_at  timestamp,
  detail        text      not null default '{}',
  foreign key (node_id) references node(id) on delete cascade
);

-- most events are over when they are written. A tracked one, such as a provisioning or a firmware update, is written while it is still running and updated as its nodes report in
create table event (
  id          text      primary key,
  kind        text      not null,
  action      text      not null,
  summary     text      not null,
  subject     text      not null default '',
  actor       text      not null,
  state       text      not null,
  severity    text      not null,
  started_at  timestamp not null,
  updated_at  timestamp not null,
  ended_at    timestamp,
  detail      text      not null default '{}'
);

create index event_started on event(started_at desc, id desc);
create index event_state on event(state);

-- nodes that start provisioning join the provisioning already under way, so there is never more than one open
create unique index event_open_provision on event(kind) where kind = 'provision' and state = 'running';

-- how an event went on one node. The node is kept by name, so an event still says what happened to a node since deleted
create table event_node (
  event_id    text      not null,
  node        text      not null,
  state       text      not null,
  stage       text      not null default '',
  message     text      not null default '',
  -- ref is what the event follows the node by, such as the id of a firmware update's job on its BMC
  ref         text      not null default '',
  updated_at  timestamp not null,
  primary key (event_id, node),
  foreign key (event_id) references event(id) on delete cascade
);

create index event_node_node on event_node(node);

-- every report of a node finishing asks whether its event still has a node going, which this answers without reading the event's other rows
create index event_node_open on event_node(event_id) where state in ('pending', 'running');

create table setting (
  key text primary key not null,
  value text not null check (json_valid(value)),
  updated_at datetime not null default current_timestamp,
  updated_by text not null default ''
) without rowid;

-- bumped with every change to setting, so a process can tell cheaply that another one changed the settings
create table setting_revision (
  id integer primary key check (id = 1),
  revision integer not null
);
insert into setting_revision (id, revision) values (1, 0);

insert into permission(method, path) values
  ('GET', '/v1/bmc/status'),
  ('GET', '/v1/bmc/sel'),
  ('GET', '/v1/bmc/firmware/updates'),
  ('POST', '/v1/bmc/power'),
  ('POST', '/v1/bmc/indicator'),
  ('POST', '/v1/bmc/jobs/clear'),
  ('POST', '/v1/bmc/sel/clear'),
  ('POST', '/v1/bmc/firmware/install'),
  ('GET', '/v1/bmc/bios'),
  ('GET', '/v1/bmc/bios/registry'),
  ('POST', '/v1/bmc/bios'),
  ('POST', '/v1/bmc/bios/discard'),
  ('GET', '/v1/bmc/identity/defaults'),
  ('POST', '/v1/bmc/asset-tag'),
  ('POST', '/v1/bmc/location'),
  ('GET', '/v1/bmc/console'),
  ('GET', '/v1/nodes/provision-status'),
  ('GET', '/v1/nodes/tags'),
  ('GET', '/v1/bmc/secureboot'),
  ('POST', '/v1/bmc/secureboot'),
  ('POST', '/v1/bmc/secureboot/keys'),
  ('POST', '/v1/bmc/secureboot/keys/delete'),
  ('POST', '/v1/bmc/secureboot/reset'),
  ('GET', '/v1/grendel/events/%'), -- :id
  ('GET', '/v1/nodes/ips/next'),
  ('GET', '/v1/nodes/ips/ranges'),
  ('POST', '/v1/nodes/import'),
  ('GET', '/v1/switch/scan'),
  ('GET', '/v1/bmc/nics'),
  ('GET', '/v1/nodes/render/%'), -- :template
  ('GET', '/v1/nodes/check'),
  ('GET', '/v1/settings'),
  ('PATCH', '/v1/settings')
;

-- every role, custom ones included, keeps what it could do: a grant on a replaced route becomes a grant on its replacement, and a new route goes to the roles holding the one it most resembles
with carried(old_method, old_path, new_method, new_path) as
(
  values
    ('GET', '/v1/bmc', 'GET', '/v1/bmc/status'),
    ('GET', '/v1/bmc', 'GET', '/v1/bmc/sel'),
    ('GET', '/v1/bmc', 'GET', '/v1/bmc/firmware/updates'),
    ('GET', '/v1/bmc', 'GET', '/v1/bmc/nics'),
    ('POST', '/v1/bmc/power/os', 'POST', '/v1/bmc/power'),
    ('POST', '/v1/bmc/power/os', 'POST', '/v1/bmc/indicator'),
    ('POST', '/v1/bmc/power/bmc', 'POST', '/v1/bmc/power'),
    ('DELETE', '/v1/bmc/sel', 'POST', '/v1/bmc/sel/clear'),
    ('DELETE', '/v1/bmc/jobs', 'POST', '/v1/bmc/jobs/clear'),
    ('DELETE', '/v1/bmc/jobs/%', 'POST', '/v1/bmc/jobs/clear'),
    ('GET', '/v1/bmc/upgrade/dell/repo', 'GET', '/v1/bmc/firmware/updates'),
    ('POST', '/v1/bmc/upgrade/dell/installfromrepo', 'POST', '/v1/bmc/firmware/install'),
    ('GET', '/v1/grendel/events', 'GET', '/v1/grendel/events/%'),
    ('POST', '/v1/nodes', 'POST', '/v1/nodes/import'),
    ('GET', '/v1/switch/%/lldp', 'GET', '/v1/switch/scan'),
    -- a render carries a boot token, so it goes to the roles that can already create one, and the check goes with it
    ('GET', '/v1/nodes/token/%', 'GET', '/v1/nodes/render/%'),
    ('GET', '/v1/nodes/token/%', 'GET', '/v1/nodes/check')
)
insert or ignore into role_permission(role_id, permission_id)
select rp.role_id, new.id
from carried
join permission old on (old.method, old.path) = (carried.old_method, carried.old_path)
join role_permission rp on rp.permission_id = old.id
join permission new on (new.method, new.path) = (carried.new_method, carried.new_path)
;

-- no route answers at these paths
delete from permission where (method, path) in
(
  ('GET', '/v1/bmc'),
  ('POST', '/v1/bmc/power/os'),
  ('POST', '/v1/bmc/power/bmc'),
  ('DELETE', '/v1/bmc/sel'),
  ('DELETE', '/v1/bmc/jobs'),
  ('DELETE', '/v1/bmc/jobs/%'),
  ('GET', '/v1/bmc/upgrade/dell/repo'),
  ('POST', '/v1/bmc/upgrade/dell/installfromrepo'),
  ('GET', '/v1/bmc/metrics'),
  ('GET', '/v1/switch/%/hosttree')
)
;

-- the foreign key removes a deleted permission's grants only where it is enforced, which the migrate command line tool does not do
delete from role_permission where permission_id not in (select id from permission);

-- creating accounts is user management, which is the admin's
delete from role_permission
where role_id in (select id from role where name = 'user')
  and permission_id in (select id from permission where (method, path) = ('POST', '/v1/users'))
;

-- the server lets admin call every route without a grant, but a role made from admin copies only its stored grants, so admin is given every permission
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name = 'admin'
;

-- read-only holds every GET that neither exposes secrets nor opens an interactive session. The role list is one of them because any role may create tokens, and picking a token's role needs it
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name in ('user', 'read-only')
  and (permission.method, permission.path) in
  (
    ('GET', '/v1/bmc/bios'),
    ('GET', '/v1/bmc/bios/registry'),
    ('GET', '/v1/bmc/identity/defaults'),
    ('GET', '/v1/bmc/secureboot'),
    ('GET', '/v1/nodes/provision-status'),
    ('GET', '/v1/nodes/tags'),
    ('GET', '/v1/nodes/ips/next'),
    ('GET', '/v1/nodes/ips/ranges'),
    ('GET', '/v1/roles')
  )
;

insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name = 'read-only'
  and (permission.method, permission.path) in
  (
    ('GET', '/v1/switch/%/lldp'),
    ('GET', '/v1/switch/scan')
  )
;

-- the user role is kept from secure boot writes, since a wrong key or a deleted platform key can leave a node unable to boot, and from settings. The rbac migration granted it node tags as /v1/nodes/tags/:action, which matched nothing, as the permission is stored with %
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name = 'user'
  and (permission.method, permission.path) in
  (
    ('POST', '/v1/bmc/bios'),
    ('POST', '/v1/bmc/bios/discard'),
    ('POST', '/v1/bmc/asset-tag'),
    ('POST', '/v1/bmc/location'),
    ('GET', '/v1/bmc/console'),
    ('PATCH', '/v1/nodes/tags/%')
  )
;
