/*
 * SPDX-FileCopyrightText: (C) 2019 Grendel Authors
 *
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

-- name: ConfigList :many
select * from config;

-- name: ConfigUpsert :exec
insert into config (key, value)
values (@key, @value)
on conflict (key)
do update set value = ?2;