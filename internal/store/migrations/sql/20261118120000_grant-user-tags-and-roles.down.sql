-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

delete from role_permission
where role_id in (select id from role where name = 'user')
  and permission_id in
  (
    select id
    from permission
    where (method, path) in
    (
      ('PATCH', '/v1/nodes/tags/%'),
      ('GET', '/v1/roles')
    )
  )
;
