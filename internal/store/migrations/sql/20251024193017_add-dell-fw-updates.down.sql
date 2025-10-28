-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

delete from role_permission where (role_id, permission_id) in
(
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
)
;

delete from role_permission where (role_id, permission_id) in
(
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
)
;

delete from permission where id in
(
  select id
  from permission
  where (method, path) in
  (
    ('GET', '/v1/bmc/upgrade/dell/repo'),
    ('POST', '/v1/bmc/upgrade/dell/installfromrepo'),
    ('DELETE', '/v1/bmc/jobs')
  )
)
;
