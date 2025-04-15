/*
 * SPDX-FileCopyrightText: (C) 2019 Grendel Authors
 *
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

-- name: RoleFetchByMethodAndPath :many
select role.name
from permission
inner join role on role_permission.role_id = role.id
inner join role_permission on role_permission.permission_id = permission.id
where permission.method = @method and @path like permission.path;

-- name: RoleFetchView :many
select permission_json from role_view;

-- name: RoleFetchViewByName :one
select permission_json from role_view where name = @name;

-- name: RoleAdd :one
insert into role (name)
values (@name)
returning id;

-- name: RoleDelete :exec
delete from role where name = @name;

-- name: RoleFetchPermissions :many
select method, path from permission;

-- name: RoleFetchPermissionsByRole :many
select permission_id from role_permission where role_id = @role_id;

-- name: RoleFetchId :one
select id from role where name = @name;

-- name: RoleFetchPermissionId :one
select id from permission where method = @method and path = @path;

-- name: RoleUpsertPermission :exec
insert into role_permission (role_id, permission_id)
values (@role_id, @permission_id)
on conflict do nothing;

-- name: RoleUpsertDelete :exec
delete from role_permission where role_id = @role_id and permission_id not in (sqlc.slice(ids));