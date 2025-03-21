-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

create trigger if not exists update_user_timestamp after update on user
    begin
        update user set updated_at = current_timestamp where id = old.id;
    end;

drop view node_view;
drop view kernel_view;

create view node_view as
select
  n.id,
  n.name,
  n.uid,
  json_object(
    'id', n.id,
    'uid', n.uid,
    'name', n.name,
    'provision', n.provision,
    'boot_image', k.name,
    'firmware', n.firmware,
    'tags',
      (select json_group_array(concat_ws(':',t.key,nt.value))
       from node_tag as nt
       join tag as t
         on nt.tag_id = t.id
       where nt.node_id = n.id
      ),
    'interfaces', (
      select json_group_array(
         json_object(
           'id', nc.id,
           'ifname', nc.name,
           'fqdn', nc.fqdn,
           'vlan', nc.vlan,
           'mac', nc.mac,
           'bmc', iif(nc.nic_type == 'bmc', true, false),
           'ip', nc.ip
         ))
       from nic as nc
       where nc.node_id = n.id and nc.nic_type != 'bond'
    ),
    'bonds', (
      select json_group_array(
         json_object(
           'id', nc.id,
           'ifname', nc.name,
           'fqdn', nc.fqdn,
           'vlan', nc.vlan,
           'mac', nc.mac,
           'peers', json_extract(nc.peers, '$'),
           'bmc', iif(nc.nic_type == 'bmc', true, false),
           'ip', nc.ip
         ))
       from nic as nc
       where nc.node_id = n.id and nc.nic_type = 'bond'
    )
  ) as host_json
from
    node as n
left join kernel as k
on n.kernel_id = k.id
;

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