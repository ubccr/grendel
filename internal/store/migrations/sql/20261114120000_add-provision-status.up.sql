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

insert into permission(method, path) values
  ('GET', '/v1/nodes/provision-status')
;

insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name in ('admin', 'user', 'read-only')
  and (permission.method, permission.path) in
  (
    ('GET', '/v1/nodes/provision-status')
  )
;
