-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

drop view if exists user_view;

create table old_user (
  id            integer primary key,
  username      text    not null unique,
  role          text    not null,
  password_hash text    not null,
  enabled       integer default false not null,
  created_at    timestamp default current_timestamp not null,
  updated_at    timestamp default current_timestamp not null
);

insert into old_user (id, username, role, password_hash, enabled, created_at, updated_at)
select user.id, user.username, role.name, user.password_hash, user.enabled, user.created_at, user.updated_at
from user
inner join role
on role.id = user.role_id;

drop table user;
alter table old_user rename to user;

drop table if exists role;
drop table if exists permission;
drop table if exists role_permission;
drop view if exists role_view;