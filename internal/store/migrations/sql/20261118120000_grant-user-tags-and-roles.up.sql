-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

-- the user role was meant to edit node tags, but the rbac migration granted it as /v1/nodes/tags/:action while the permission is stored as /v1/nodes/tags/%, so the grant matched nothing. Users may also list roles, so they can see which roles a token of theirs can carry
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name = 'user'
  and (permission.method, permission.path) in
  (
    ('PATCH', '/v1/nodes/tags/%'),
    ('GET', '/v1/roles')
  )
;
