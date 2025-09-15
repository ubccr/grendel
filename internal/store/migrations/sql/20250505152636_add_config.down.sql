-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

drop table if exists config;

delete from role_permission
where (permission_id) in
(
    select id
    from permission
    where (method, path) in
    (
        ('GET', '/v1/config/get/file'),
        ('GET', '/v1/config/get'),
        ('PATCH', '/v1/config/set')
    )
);

delete from permission
where (method, path) in
    (
        ('GET', '/v1/config/get/file'),
        ('GET', '/v1/config/get'),
        ('PATCH', '/v1/config/set')
    )
;