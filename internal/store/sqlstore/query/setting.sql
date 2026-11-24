/*
 * SPDX-FileCopyrightText: (C) 2019 Grendel Authors
 *
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

-- name: SettingList :many
select key, value, updated_at, updated_by
from setting
order by key;

-- name: SettingUpsert :exec
insert into setting (key, value, updated_at, updated_by)
values (@key, @value, current_timestamp, @updated_by)
on conflict (key)
do update set
  value = excluded.value,
  updated_at = excluded.updated_at,
  updated_by = excluded.updated_by;

-- name: SettingDelete :exec
delete from setting where key = @key;

-- name: SettingInsert :exec
insert into setting (key, value, updated_at, updated_by)
values (@key, @value, current_timestamp, @updated_by)
on conflict (key) do nothing;

-- name: SettingRevision :one
select revision from setting_revision where id = 1;

-- name: SettingRevisionBump :one
update setting_revision set revision = revision + 1 where id = 1 returning revision;
