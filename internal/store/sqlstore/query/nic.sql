/*
 * SPDX-FileCopyrightText: (C) 2019 Grendel Authors
 *
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

-- name: NicUpsert :one
insert into nic (id, node_id, nic_type, name, vlan, fqdn, mac, ip, peers, mtu)
values (sqlc.narg(id), @node_id, @nic_type, @name, @vlan, @fqdn, @mac, @ip, @peers, @mtu)
on conflict (id)
do update set nic_type = ?3, name = ?4, vlan = ?5, fqdn = ?6, mac = ?7, ip = ?8,
              peers = ?9, mtu = ?10
returning *;

-- name: NicUpsertDelete :exec
delete from nic where node_id = @node_id and id not in (sqlc.slice(ids));
