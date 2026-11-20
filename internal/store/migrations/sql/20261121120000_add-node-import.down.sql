-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

delete from role_permission where permission_id in
(
  select id
  from permission
  where (method, path) in
  (
    ('POST', '/v1/nodes/import'),
    ('GET', '/v1/switch/scan'),
    ('GET', '/v1/bmc/nics')
  )
)
;

delete from permission where (method, path) in
(
  ('POST', '/v1/nodes/import'),
  ('GET', '/v1/switch/scan'),
  ('GET', '/v1/bmc/nics')
)
;
