-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

insert into permission(method, path) values
  ('GET', '/v1/bmc/secureboot'),
  ('POST', '/v1/bmc/secureboot'),
  ('POST', '/v1/bmc/secureboot/keys'),
  ('POST', '/v1/bmc/secureboot/keys/delete'),
  ('POST', '/v1/bmc/secureboot/reset')
;

-- the read, which includes the keys, goes to every built in role. Keys are public certificates and hashes, so they are no secret
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name in ('admin', 'user', 'read-only')
  and (permission.method, permission.path) in
  (
    ('GET', '/v1/bmc/secureboot')
  )
;

-- the writes are admin only, since a wrong key or a deleted platform key can leave a node unable to boot
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name = 'admin'
  and (permission.method, permission.path) in
  (
    ('POST', '/v1/bmc/secureboot'),
    ('POST', '/v1/bmc/secureboot/keys'),
    ('POST', '/v1/bmc/secureboot/keys/delete'),
    ('POST', '/v1/bmc/secureboot/reset')
  )
;
