-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

insert into permission(method, path) values
  ('POST', '/v1/nodes/import'),
  ('GET', '/v1/switch/scan'),
  ('GET', '/v1/bmc/nics')
;

-- each goes to the roles already holding the permission it most resembles, so custom roles keep the access they were given: an import saves nodes, a scan reads switches as LLDP does, and listing NICs reads BMCs as status does
insert or ignore into role_permission(role_id, permission_id)
select rp.role_id, new.id
from role_permission rp
join permission old on old.id = rp.permission_id
join permission new on (old.method, old.path, new.method, new.path) in
  (
    ('POST', '/v1/nodes', 'POST', '/v1/nodes/import'),
    ('GET', '/v1/switch/%/lldp', 'GET', '/v1/switch/scan'),
    ('GET', '/v1/bmc/status', 'GET', '/v1/bmc/nics')
  )
;
