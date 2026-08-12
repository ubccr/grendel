-- SPDX-FileCopyrightText: (C) 2019 Grendel Authors
--
-- SPDX-License-Identifier: GPL-3.0-or-later

-- An interface FQDN is now exactly one name, no comma separation.
UPDATE nic
SET
  fqdn = trim(substr(fqdn, 1, instr(fqdn, ',') - 1))
WHERE
  instr(coalesce(fqdn, ''), ',') > 0;

-- Lookups are exact now, and lower('') = '' would make an empty column match a query for the empty name. Store the absence of a name as NULL instead.
UPDATE nic SET fqdn = null WHERE trim(coalesce(fqdn, '')) = '';

CREATE INDEX IF NOT EXISTS nic_fqdn_lower_idx ON nic (lower(fqdn));

-- Split the CIDR that nic.ip used to hold into the address and its prefix length, so nic.ip means the IP address and can be indexed directly.
UPDATE nic SET ip = null WHERE trim(coalesce(ip, '')) = '';

ALTER TABLE nic ADD COLUMN prefix_len integer;

UPDATE nic
SET
  prefix_len = CAST(substr(ip, instr(ip, '/') + 1) AS integer),
  ip = substr(ip, 1, instr(ip, '/') - 1)
WHERE
  instr(coalesce(ip, ''), '/') > 0;

-- An address stored without a mask becomes an explicit host route.
UPDATE nic
SET
  prefix_len = 32
WHERE
  ip IS NOT null
  AND prefix_len IS null;

CREATE INDEX IF NOT EXISTS nic_ip_idx ON nic (ip);

CREATE INDEX IF NOT EXISTS nic_mac_idx ON nic (mac);

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
