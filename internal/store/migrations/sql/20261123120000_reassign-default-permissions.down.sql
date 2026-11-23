-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

delete from role_permission
where role_id in (select id from role where name = 'read-only')
  and permission_id in
  (
    select id
    from permission
    where (method, path) in
    (
      ('GET', '/v1/nodes/ips/next'),
      ('GET', '/v1/switch/%/lldp'),
      ('GET', '/v1/switch/scan'),
      ('GET', '/v1/roles')
    )
  )
;

insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name = 'user'
  and (permission.method, permission.path) in
  (
    ('POST', '/v1/users')
  )
;

insert into permission(method, path) values
  ('GET', '/v1/switch/%/hosttree')
;

insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name in ('admin', 'user')
  and (permission.method, permission.path) in
  (
    ('GET', '/v1/switch/%/hosttree')
  )
;
