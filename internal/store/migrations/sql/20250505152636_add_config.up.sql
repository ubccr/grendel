-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

create table if not exists config (
    key text primary key,
    value text
);

insert into permission(method, path) values
  ('GET', '/v1/config/get/file'),
  ('GET', '/v1/config/get'),
  ('PATCH', '/v1/config/set')
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
    where (method, path) in
      (
        ('GET', '/v1/config/get/file'),
        ('GET', '/v1/config/get'),
        ('PATCH', '/v1/config/set')
      )
  ) permission
;