-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

delete from role_permission where permission_id in
(
  select id
  from permission
  where (method, path) in
  (
    ('GET', '/v1/bmc/bios'),
    ('GET', '/v1/bmc/bios/registry'),
    ('POST', '/v1/bmc/bios'),
    ('POST', '/v1/bmc/bios/discard')
  )
)
;

delete from permission where (method, path) in
(
  ('GET', '/v1/bmc/bios'),
  ('GET', '/v1/bmc/bios/registry'),
  ('POST', '/v1/bmc/bios'),
  ('POST', '/v1/bmc/bios/discard')
)
;
