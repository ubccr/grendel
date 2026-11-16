-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

delete from role_permission where permission_id in
(
  select id
  from permission
  where (method, path) in
  (
    ('GET', '/v1/grendel/events/%')
  )
)
;

delete from permission where (method, path) in
(
  ('GET', '/v1/grendel/events/%')
)
;

drop table event_node;
drop table event;
