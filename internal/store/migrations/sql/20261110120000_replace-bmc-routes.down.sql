-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

insert into permission(method, path) values
  ('GET', '/v1/bmc'),
  ('POST', '/v1/bmc/power/os'),
  ('POST', '/v1/bmc/power/bmc'),
  ('DELETE', '/v1/bmc/sel'),
  ('DELETE', '/v1/bmc/jobs'),
  ('DELETE', '/v1/bmc/jobs/%'),
  ('GET', '/v1/bmc/upgrade/dell/repo'),
  ('POST', '/v1/bmc/upgrade/dell/installfromrepo'),
  ('GET', '/v1/bmc/metrics')
;

-- the old routes get back the roles that hold their replacement. The repo listing follows the firmware install rather than the firmware updates it became, since every role that could read a BMC was given those and only some could list the repo
with replaced(old_method, old_path, new_method, new_path) as
(
  values
    ('GET', '/v1/bmc', 'GET', '/v1/bmc/status'),
    ('POST', '/v1/bmc/power/os', 'POST', '/v1/bmc/power'),
    ('POST', '/v1/bmc/power/bmc', 'POST', '/v1/bmc/power'),
    ('DELETE', '/v1/bmc/sel', 'POST', '/v1/bmc/sel/clear'),
    ('DELETE', '/v1/bmc/jobs', 'POST', '/v1/bmc/jobs/clear'),
    ('DELETE', '/v1/bmc/jobs/%', 'POST', '/v1/bmc/jobs/clear'),
    ('GET', '/v1/bmc/upgrade/dell/repo', 'POST', '/v1/bmc/firmware/install'),
    ('POST', '/v1/bmc/upgrade/dell/installfromrepo', 'POST', '/v1/bmc/firmware/install')
)
insert or ignore into role_permission(role_id, permission_id)
select rp.role_id, old.id
from replaced
join permission new on (new.method, new.path) = (replaced.new_method, replaced.new_path)
join role_permission rp on rp.permission_id = new.id
join permission old on (old.method, old.path) = (replaced.old_method, replaced.old_path)
;

-- metrics was removed with nothing replacing it, so which roles held it is not known. Only admin, which holds every permission, gets it back
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name = 'admin'
  and (permission.method, permission.path) = ('GET', '/v1/bmc/metrics')
;

delete from role_permission where permission_id in
(
  select id
  from permission
  where (method, path) in
  (
    ('GET', '/v1/bmc/status'),
    ('GET', '/v1/bmc/sel'),
    ('GET', '/v1/bmc/firmware/updates'),
    ('POST', '/v1/bmc/power'),
    ('POST', '/v1/bmc/indicator'),
    ('POST', '/v1/bmc/jobs/clear'),
    ('POST', '/v1/bmc/sel/clear'),
    ('POST', '/v1/bmc/firmware/install')
  )
)
;

delete from permission where (method, path) in
(
  ('GET', '/v1/bmc/status'),
  ('GET', '/v1/bmc/sel'),
  ('GET', '/v1/bmc/firmware/updates'),
  ('POST', '/v1/bmc/power'),
  ('POST', '/v1/bmc/indicator'),
  ('POST', '/v1/bmc/jobs/clear'),
  ('POST', '/v1/bmc/sel/clear'),
  ('POST', '/v1/bmc/firmware/install')
)
;
