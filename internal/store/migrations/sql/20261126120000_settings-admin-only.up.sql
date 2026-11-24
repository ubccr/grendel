-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

-- the server lets admin call every route without a grant, but a role made from admin copies only its stored grants, so admin needs the rows too
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name = 'admin'
  and (permission.method, permission.path) in
  (
    ('GET', '/v1/settings'),
    ('PATCH', '/v1/settings')
  )
;

-- settings are the admin's to read as well as to change
delete from role_permission
where role_id in (select id from role where name in ('user', 'read-only'))
  and permission_id in (select id from permission where (method, path) in (('GET', '/v1/settings')))
;
