-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

-- an event is something that was done to grendel or its nodes. Most are over when they are written. A tracked event, such as a provisioning or a firmware update, is written while it is still running and updated as its nodes report in
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
  detail      text      not null default '{}',
  -- counts are the event's nodes by state as of when it was last settled, so listing finished events does not count their nodes again. A running event's are counted from its nodes
  counts      text      not null default '{}'
);

-- events are listed newest first
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

insert into permission(method, path) values
  ('GET', '/v1/grendel/events/%')
;

insert or ignore into role_permission(role_id, permission_id)
select rp.role_id, new.id
from role_permission as rp
join permission as old on old.id = rp.permission_id
join permission as new on (new.method, new.path) = ('GET', '/v1/grendel/events/%')
where (old.method, old.path) = ('GET', '/v1/grendel/events')
;
