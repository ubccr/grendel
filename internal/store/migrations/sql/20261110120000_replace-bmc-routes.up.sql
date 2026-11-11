-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

-- GET /v1/bmc/jobs and the configure routes keep their paths and grants
insert into permission(method, path) values
  ('GET', '/v1/bmc/status'),
  ('GET', '/v1/bmc/sel'),
  ('GET', '/v1/bmc/firmware/updates'),
  ('POST', '/v1/bmc/power'),
  ('POST', '/v1/bmc/indicator'),
  ('POST', '/v1/bmc/jobs/clear'),
  ('POST', '/v1/bmc/sel/clear'),
  ('POST', '/v1/bmc/firmware/install')
;

-- every role, custom ones included, keeps what it could do: a grant on a replaced route becomes a grant on its replacement
with replaced(old_method, old_path, new_method, new_path) as
(
  values
    ('GET', '/v1/bmc', 'GET', '/v1/bmc/status'),
    ('GET', '/v1/bmc', 'GET', '/v1/bmc/sel'),
    ('POST', '/v1/bmc/power/os', 'POST', '/v1/bmc/power'),
    ('POST', '/v1/bmc/power/os', 'POST', '/v1/bmc/indicator'),
    ('POST', '/v1/bmc/power/bmc', 'POST', '/v1/bmc/power'),
    ('DELETE', '/v1/bmc/sel', 'POST', '/v1/bmc/sel/clear'),
    ('DELETE', '/v1/bmc/jobs', 'POST', '/v1/bmc/jobs/clear'),
    ('DELETE', '/v1/bmc/jobs/%', 'POST', '/v1/bmc/jobs/clear'),
    ('GET', '/v1/bmc/upgrade/dell/repo', 'GET', '/v1/bmc/firmware/updates'),
    ('POST', '/v1/bmc/upgrade/dell/installfromrepo', 'POST', '/v1/bmc/firmware/install')
)
insert or ignore into role_permission(role_id, permission_id)
select rp.role_id, new.id
from replaced
join permission old on (old.method, old.path) = (replaced.old_method, replaced.old_path)
join role_permission rp on rp.permission_id = old.id
join permission new on (new.method, new.path) = (replaced.new_method, replaced.new_path)
;

-- the reads go to every built in role. The admin grant in the rbac migration selected every permission that existed at the time, so a permission added later has to name admin itself
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name in ('admin', 'user', 'read-only')
  and (permission.method, permission.path) in
  (
    ('GET', '/v1/bmc/status'),
    ('GET', '/v1/bmc/sel'),
    ('GET', '/v1/bmc/firmware/updates')
  )
;

-- the writes stop short of read-only
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where role.name in ('admin', 'user')
  and (permission.method, permission.path) in
  (
    ('POST', '/v1/bmc/power'),
    ('POST', '/v1/bmc/indicator'),
    ('POST', '/v1/bmc/jobs/clear'),
    ('POST', '/v1/bmc/sel/clear')
  )
;

-- the replaced routes, and metrics, which is removed outright
delete from role_permission where permission_id in
(
  select id
  from permission
  where (method, path) in
  (
    ('GET', '/v1/bmc'),
    ('POST', '/v1/bmc/power/os'),
    ('POST', '/v1/bmc/power/bmc'),
    ('DELETE', '/v1/bmc/sel'),
    ('DELETE', '/v1/bmc/jobs'),
    ('DELETE', '/v1/bmc/jobs/%'),
    ('GET', '/v1/bmc/upgrade/dell/repo'),
    ('POST', '/v1/bmc/upgrade/dell/installfromrepo'),
    ('GET', '/v1/bmc/metrics')
  )
)
;

delete from permission where (method, path) in
(
  ('GET', '/v1/bmc'),
  ('POST', '/v1/bmc/power/os'),
  ('POST', '/v1/bmc/power/bmc'),
  ('DELETE', '/v1/bmc/sel'),
  ('DELETE', '/v1/bmc/jobs'),
  ('DELETE', '/v1/bmc/jobs/%'),
  ('GET', '/v1/bmc/upgrade/dell/repo'),
  ('POST', '/v1/bmc/upgrade/dell/installfromrepo'),
  ('GET', '/v1/bmc/metrics')
)
;
