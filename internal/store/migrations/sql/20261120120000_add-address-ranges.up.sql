-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

insert into permission(method, path) values
  ('GET', '/v1/nodes/ips/ranges')
;

-- listing where nodes' addresses are only reads what listing the nodes already shows, so it goes to every built in role
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name in ('admin', 'user', 'read-only')
  and (permission.method, permission.path) in
  (
    ('GET', '/v1/nodes/ips/ranges')
  )
;
