-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

create table if not exists role (
  id            integer primary key,
  name          text    not null unique
);

create table if not exists permission (
  id      integer primary key,
  method  text    not null,
  path    text    not null
);

create table if not exists role_permission (
  role_id       integer not null,
  permission_id integer not null,
  foreign key (role_id) references role(id) on delete cascade,
  foreign key (permission_id) references permission(id) on delete cascade,
  unique (role_id, permission_id)
);

insert into permission(method, path) values
  ('GET', '/v1/db/dump'),
  ('GET', '/v1/users'),
  ('GET', '/v1/grendel/events'),
  ('GET', '/v1/bmc'),
  ('GET', '/v1/nodes/token/%'), -- :interface
  ('GET', '/v1/nodes/find'),
  ('GET', '/v1/bmc/jobs'),
  ('GET', '/v1/bmc/metrics'),
  ('GET', '/v1/nodes'),
  ('GET', '/v1/images'),
  ('GET', '/v1/images/find'),
  ('GET', '/v1/roles'),
  ('PATCH', '/v1/nodes/provision'),
  ('PATCH', '/v1/nodes/tags/%'), -- :action
  ('PATCH', '/v1/nodes/image'),
  ('PATCH', '/v1/users/%/role'), -- :usernames
  ('PATCH', '/v1/users/%/enable'), -- :usernames
  ('PATCH', '/v1/roles'),
  ('PATCH', '/v1/auth/reset'),
  ('POST', '/v1/users'),
  ('POST', '/v1/bmc/power/os'),
  ('POST', '/v1/bmc/power/bmc'),
  ('POST', '/v1/bmc/configure/auto'),
  ('POST', '/v1/db/restore'),
  ('POST', '/v1/nodes'),
  ('POST', '/v1/images'),
  ('POST', '/v1/auth/token'),
  ('POST', '/v1/bmc/configure/import'),
  ('POST', '/v1/roles'),
  ('DELETE', '/v1/bmc/jobs/%'), -- :jids
  ('DELETE', '/v1/users/%'), -- :usernames
  ('DELETE', '/v1/bmc/sel'),
  ('DELETE', '/v1/nodes'),
  ('DELETE', '/v1/images'),
  ('DELETE', '/v1/roles/%') -- :names
;

insert into role(name) values
  ('admin'),
  ('user'),
  ('read-only')
;

insert into role_permission(role_id, permission_id)
select role.id, permission.id
from
  (
    select id
    from role
    where name = 'admin'
  ) role,
  (
    select id
    from permission
  ) permission
;

insert into role_permission(role_id, permission_id)
select role.id, permission.id
from
  (
    select id
    from role
    where name = 'user'
  ) role,
  (
    select id
    from permission
    where (method, path) in
      (
        ('GET', '/v1/grendel/events'),
        ('GET', '/v1/bmc'),
        ('GET', '/v1/nodes/find'),
        ('GET', '/v1/bmc/jobs'),
        ('GET', '/v1/bmc/metrics'),
        ('GET', '/v1/nodes'),
        ('GET', '/v1/images'),
        ('GET', '/v1/images/find'),
        ('PATCH', '/v1/nodes/provision'),
        ('PATCH', '/v1/nodes/tags/:action'),
        ('PATCH', '/v1/nodes/image'),
        ('PATCH', '/v1/auth/reset'),
        ('POST', '/v1/users'),
        ('POST', '/v1/bmc/power/os'),
        ('POST', '/v1/bmc/power/bmc'),
        ('POST', '/v1/bmc/configure/auto'),
        ('POST', '/v1/nodes'),
        ('POST', '/v1/images'),
        ('POST', '/v1/auth/token'),
        ('POST', '/v1/bmc/configure/import'),
        ('DELETE', '/v1/bmc/jobs/:jids'),
        ('DELETE', '/v1/bmc/sel'),
        ('DELETE', '/v1/nodes'),
        ('DELETE', '/v1/images')
      )
  ) permission
;

insert into role_permission(role_id, permission_id)
select role.id, permission.id
from
  (
    select id
    from role
    where name = 'read-only'
  ) role,
  (
    select id
    from permission
    where (method, path) in
      (
        ('GET', '/v1/grendel/events'),
        ('GET', '/v1/bmc'),
        ('GET', '/v1/nodes/find'),
        ('GET', '/v1/bmc/jobs'),
        ('GET', '/v1/bmc/metrics'),
        ('GET', '/v1/nodes'),
        ('GET', '/v1/images'),
        ('GET', '/v1/images/find'),
        ('PATCH', '/v1/auth/reset'),
        ('POST', '/v1/auth/token')
      )
  ) permission
;

create view role_view as
  select
  role.name,
  json_object(
    'name', role.name,
    'permission_list', (
      select json_group_array(
        json_object(
        'method', permission.method,
        'path', permission.path
        )
      )
      from role_permission
      inner join permission
      on role_permission.permission_id = permission.id
      where role_permission.role_id = role.id
    ),
    'unassigned_permission_list', (
      select json_group_array(
        json_object(
        'method', permission.method,
        'path', permission.path
        )
      )
      from permission
      where not exists
      (
        select * from role_permission
        where role_permission.role_id = role.id
        and role_permission.permission_id = permission.id
      )
    )
  ) as permission_json
  from
    role
;

-- set at least one user to enabled since we were not setting this field pre v0.2.1
update user
set enabled = 1
where role = 'admin' 
and not exists
(
  select * from user where enabled = 1
);

-- update user table
create table new_user (
  id            integer primary key,
  username      text    not null unique,
  role_id       integer not null,
  password_hash text    not null,
  enabled       integer default false not null,
  created_at    timestamp default current_timestamp not null,
  updated_at    timestamp default current_timestamp not null,
  foreign key (role_id) references role(id) on delete restrict
);

insert into new_user (id, username, role_id, password_hash, enabled, created_at, updated_at)
select user.id, user.username, role.id, user.password_hash, user.enabled, user.created_at, user.updated_at
from user
inner join role
on role.name = user.role;

drop table user;
alter table new_user rename to user;

create view if not exists user_view as
select
user.username,
user.id,
user.password_hash,
role.name as role,
user.enabled,
user.created_at,
user.updated_at
from user
inner join role
on role.id = user.role_id;