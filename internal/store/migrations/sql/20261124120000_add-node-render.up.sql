-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

insert into permission(method, path) values
  ('GET', '/v1/nodes/render/%'), -- :template
  ('GET', '/v1/nodes/check')
;

-- rendered output carries a boot token, so it goes to the roles that can already create one
insert or ignore into role_permission(role_id, permission_id)
select rp.role_id, new.id
from role_permission rp
join permission old on old.id = rp.permission_id
join permission new on (old.method, old.path, new.method, new.path) in
  (
    ('GET', '/v1/nodes/token/%', 'GET', '/v1/nodes/render/%'),
    ('GET', '/v1/nodes/token/%', 'GET', '/v1/nodes/check')
  )
;
