-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

-- read-only holds every GET that neither exposes secrets nor opens an interactive session. The role list is one of them because a read-only user may create tokens, and picking a token's role needs it
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name = 'read-only'
  and (permission.method, permission.path) in
  (
    ('GET', '/v1/nodes/ips/next'),
    ('GET', '/v1/switch/%/lldp'),
    ('GET', '/v1/switch/scan'),
    ('GET', '/v1/roles')
  )
;

-- creating accounts is user management, which is the admin's
delete from role_permission
where role_id in (select id from role where name = 'user')
  and permission_id in (select id from permission where (method, path) in (('POST', '/v1/users')))
;

-- no route answers at this path
delete from permission where (method, path) in (('GET', '/v1/switch/%/hosttree'));
