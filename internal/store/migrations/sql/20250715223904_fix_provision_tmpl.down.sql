-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

drop view kernel_view;

create view kernel_view as
select
  k.id,
  k.name,
  json_object(
    'id', k.id,
    'uid', k.uid,
    'name', k.name,
    'kernel', k.path,
    'cmdline', k.command_line,
    'verify', iif(k.verify == 0, json('false'), json('true')),
    'initrd', (
      select json_group_array(
         rd.path
       )
       from initrd as rd
       where rd.kernel_id = k.id
    ),
    'provision_templates', (
      select json_group_object(tt.uri_name, t.name)
      from kernel_template as kt
      join template t
         on kt.template_id = t.id
      join template_type tt
         on t.template_type_id = tt.id and tt.name != 'butane' and tt.name != 'user_data'
      where kt.kernel_id = k.id
    ),
    'butane', (
      select t.name
      from kernel_template as kt
      join template t
         on kt.template_id = t.id
      join template_type tt
         on t.template_type_id = tt.id and tt.name = 'butane'
      where kt.kernel_id = k.id
    ),
    'user_data', (
      select t.name
      from kernel_template as kt
      join template t
         on kt.template_id = t.id
      join template_type tt
         on t.template_type_id = tt.id and tt.name = 'user_data'
      where kt.kernel_id = k.id
    )
  ) as image_json
from
    kernel as k
;
