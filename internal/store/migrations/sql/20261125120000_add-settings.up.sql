-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

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
  ('GET', '/v1/settings'),
  ('PATCH', '/v1/settings')
;

-- the list never carries a secret's value, and the command line client reads the default image from it
insert or ignore into role_permission(role_id, permission_id)
select r.id, p.id
from role r, permission p
where r.name in ('user', 'read-only') and p.method = 'GET' and p.path = '/v1/settings'
;
