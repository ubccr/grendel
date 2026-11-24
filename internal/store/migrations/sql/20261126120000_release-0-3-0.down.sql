-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

insert into permission(method, path) values
  ('GET', '/v1/bmc'),
  ('POST', '/v1/bmc/power/os'),
  ('POST', '/v1/bmc/power/bmc'),
  ('DELETE', '/v1/bmc/sel'),
  ('DELETE', '/v1/bmc/jobs'),
  ('DELETE', '/v1/bmc/jobs/%'), -- :jids
  ('GET', '/v1/bmc/upgrade/dell/repo'),
  ('POST', '/v1/bmc/upgrade/dell/installfromrepo'),
  ('GET', '/v1/bmc/metrics'),
  ('GET', '/v1/switch/%/hosttree') -- :nodeset
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

-- nothing replaced metrics or the host tree, so which roles held them is not known and they go back to the built in roles that held them at first
insert or ignore into role_permission(role_id, permission_id)
select role.id, permission.id
from role, permission
where (role.name, permission.method, permission.path) in
  (
    ('admin', 'GET', '/v1/bmc/metrics'),
    ('admin', 'GET', '/v1/switch/%/hosttree'),
    ('user', 'GET', '/v1/switch/%/hosttree'),
    ('user', 'POST', '/v1/users')
  )
;

-- the user role keeps node tags and the role list. Nothing records whether it held them before the up, so leaving one the up added costs less than taking away one an admin gave
delete from role_permission
where role_id in (select id from role where name = 'read-only')
  and permission_id in
  (
    select id
    from permission
    where (method, path) in
    (
      ('GET', '/v1/switch/%/lldp'),
      ('GET', '/v1/roles')
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
  ('POST', '/v1/bmc/firmware/install'),
  ('GET', '/v1/bmc/bios'),
  ('GET', '/v1/bmc/bios/registry'),
  ('POST', '/v1/bmc/bios'),
  ('POST', '/v1/bmc/bios/discard'),
  ('GET', '/v1/bmc/identity/defaults'),
  ('POST', '/v1/bmc/asset-tag'),
  ('POST', '/v1/bmc/location'),
  ('GET', '/v1/bmc/console'),
  ('GET', '/v1/nodes/provision-status'),
  ('GET', '/v1/nodes/tags'),
  ('GET', '/v1/bmc/secureboot'),
  ('POST', '/v1/bmc/secureboot'),
  ('POST', '/v1/bmc/secureboot/keys'),
  ('POST', '/v1/bmc/secureboot/keys/delete'),
  ('POST', '/v1/bmc/secureboot/reset'),
  ('GET', '/v1/grendel/events/%'),
  ('GET', '/v1/nodes/ips/next'),
  ('GET', '/v1/nodes/ips/ranges'),
  ('POST', '/v1/nodes/import'),
  ('GET', '/v1/switch/scan'),
  ('GET', '/v1/bmc/nics'),
  ('GET', '/v1/nodes/render/%'),
  ('GET', '/v1/nodes/check'),
  ('GET', '/v1/settings'),
  ('PATCH', '/v1/settings')
)
;

-- the foreign key removes a deleted permission's grants only where it is enforced, which the migrate command line tool does not do
delete from role_permission where permission_id not in (select id from permission);

drop table setting_revision;
drop table setting;
drop table event_node;
drop table event;
drop table provision_status;
