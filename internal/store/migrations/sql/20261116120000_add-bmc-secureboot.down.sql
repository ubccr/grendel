-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

delete from role_permission where permission_id in
(
  select id
  from permission
  where (method, path) in
  (
    ('GET', '/v1/bmc/secureboot'),
    ('POST', '/v1/bmc/secureboot'),
    ('POST', '/v1/bmc/secureboot/keys'),
    ('POST', '/v1/bmc/secureboot/keys/delete'),
    ('POST', '/v1/bmc/secureboot/reset')
  )
)
;

delete from permission where (method, path) in
(
  ('GET', '/v1/bmc/secureboot'),
  ('POST', '/v1/bmc/secureboot'),
  ('POST', '/v1/bmc/secureboot/keys'),
  ('POST', '/v1/bmc/secureboot/keys/delete'),
  ('POST', '/v1/bmc/secureboot/reset')
)
;
