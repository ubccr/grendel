-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

insert into permission(method, path) values
  ('GET', '/v1/bmc/upgrade/dell/repo'),
  ('POST', '/v1/bmc/upgrade/dell/installfromrepo'),
  ('DELETE', '/v1/bmc/jobs')
;

insert into role_permission(role_id, permission_id)
select role.id, permission.id
from
  (
    select id
    from role
    where name = 'admin'
  ) role,
  (
    select id
    from permission
    where (method, path) in
      (
        ('GET', '/v1/bmc/upgrade/dell/repo'),
        ('POST', '/v1/bmc/upgrade/dell/installfromrepo'),
        ('DELETE', '/v1/bmc/jobs')
      )
  ) permission
;

insert into role_permission(role_id, permission_id)
select role.id, permission.id
from
  (
    select id
    from role
    where name = 'user'
  ) role,
  (
    select id
    from permission
    where (method, path) in
      (
        ('GET', '/v1/bmc/upgrade/dell/repo'),
        ('POST', '/v1/bmc/upgrade/dell/installfromrepo'),
        ('DELETE', '/v1/bmc/jobs')
      )
  ) permission
;
