-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

insert into permission(method, path) values
  ('GET', '/v1/bmc/identity/defaults'),
  ('POST', '/v1/bmc/asset-tag'),
  ('POST', '/v1/bmc/location')
;

-- the reads go to every built in role. The admin grant in the rbac migration selected every permission that existed at the time, so a permission added later has to name admin itself
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name in ('admin', 'user', 'read-only')
  and (permission.method, permission.path) in
  (
    ('GET', '/v1/bmc/identity/defaults')
  )
;

-- the writes stop short of read-only
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name in ('admin', 'user')
  and (permission.method, permission.path) in
  (
    ('POST', '/v1/bmc/asset-tag'),
    ('POST', '/v1/bmc/location')
  )
;
