-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

delete from role_permission where permission_id in
(
  select id
  from permission
  where (method, path) in
  (
    ('GET', '/v1/settings'),
    ('PATCH', '/v1/settings')
  )
)
;

delete from permission where (method, path) in
(
  ('GET', '/v1/settings'),
  ('PATCH', '/v1/settings')
)
;

drop table setting_revision;
drop table setting;
