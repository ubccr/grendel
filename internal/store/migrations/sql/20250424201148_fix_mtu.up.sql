-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

drop view node_view;

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
           'mtu', nc.mtu,
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
           'mtu', nc.mtu,
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