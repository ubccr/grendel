-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

DROP VIEW node_view;

CREATE VIEW node_view AS
SELECT
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
    (
      SELECT json_group_array(t.key)
      FROM node_tag AS nt
      JOIN tag AS t
        ON nt.tag_id = t.id
      WHERE
        nt.node_id = n.id
    ),
    'interfaces', (
      SELECT
        json_group_array(
          json_object(
            'id', nc.id,
            'ifname', nc.name,
            'fqdn', nc.fqdn,
            'vlan', nc.vlan,
            'mac', nc.mac,
            'mtu', nc.mtu,
            'bmc', iif(nc.nic_type = 'bmc', true, false),
            -- Recombined so model.NetInterface.IP can netip.ParsePrefix it.
            'ip', iif(nc.ip IS null, null, nc.ip || '/' || coalesce(nc.prefix_len, iif(instr(nc.ip, ':') > 0, 128, 32))
            )
          ))
      FROM nic AS nc
      WHERE nc.node_id = n.id AND nc.nic_type != 'bond'
    ),
    'bonds', (
      SELECT
        json_group_array(
          json_object(
            'id', nc.id,
            'ifname', nc.name,
            'fqdn', nc.fqdn,
            'vlan', nc.vlan,
            'mac', nc.mac,
            'peers', json_extract(nc.peers, '$'),
            'mtu', nc.mtu,
            'bmc', iif(nc.nic_type = 'bmc', true, false),
            -- Recombined so model.NetInterface.IP can netip.ParsePrefix it.
            'ip', iif(nc.ip IS null, null, nc.ip || '/' || coalesce(nc.prefix_len, iif(instr(nc.ip, ':') > 0, 128, 32))
            )
          ))
      FROM nic AS nc
      WHERE
        nc.node_id = n.id AND nc.nic_type = 'bond'
    )
  ) AS host_json
FROM node AS n
LEFT JOIN kernel AS k
  ON n.kernel_id = k.id;

DROP INDEX IF EXISTS arch_name_idx;

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
         on t.template_type_id = tt.id
      where kt.kernel_id = k.id
    )
  ) as image_json
from
    kernel as k
;
