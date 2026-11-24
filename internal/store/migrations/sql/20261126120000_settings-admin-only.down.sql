-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

-- a built in role cannot be edited, so each of these holds exactly what the migration before this one left it
delete from role_permission
where role_id in (select id from role where name = 'admin')
  and permission_id in
  (
    select id
    from permission
    where (method, path) in
    (
      ('GET', '/v1/settings'),
      ('PATCH', '/v1/settings')
    )
  )
;

insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name in ('user', 'read-only')
  and (permission.method, permission.path) = ('GET', '/v1/settings')
;
