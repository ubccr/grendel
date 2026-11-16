-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

insert into permission(method, path) values
  ('GET', '/v1/nodes/ips/next')
;

-- suggesting an address is for adding or editing a node, so it goes to the roles that can save one
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name in ('admin', 'user')
  and (permission.method, permission.path) in
  (
    ('GET', '/v1/nodes/ips/next')
  )
;
