-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

insert into permission(method, path) values
  ('GET', '/v1/switch/%/hosttree'),
  ('GET', '/v1/switch/%/lldp')
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
        ('GET', '/v1/switch/%/hosttree'),
        ('GET', '/v1/switch/%/lldp')
      )
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
        ('GET', '/v1/switch/%/hosttree'),
        ('GET', '/v1/switch/%/lldp')
      )
  ) permission
;