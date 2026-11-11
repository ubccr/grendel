-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

insert into permission(method, path) values
  ('GET', '/v1/bmc/console')
;

-- a console takes keystrokes, so it stops short of read-only
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name in ('admin', 'user')
  and (permission.method, permission.path) in
  (
    ('GET', '/v1/bmc/console')
  )
;
