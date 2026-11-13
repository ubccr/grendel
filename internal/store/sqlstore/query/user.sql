/*
 * SPDX-FileCopyrightText: (C) 2019 Grendel Authors
 *
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

-- name: UserCount :one
select count(*) from user;

-- name: UserFetch :one
select * from user_view where username = @username;

-- name: UserList :many
select * from user_view;

-- name: UserCreate :one
insert into user (username, password_hash, role_id, enabled)
select @username, @password_hash, role.id, @enabled
from role
where role.name = @role
returning *;

-- name: UserRestore :one
insert into user (username, password_hash, role_id, enabled)
select @username, @password_hash, role.id, @enabled
from role
where role.name = @role
on conflict (username)
do update set password_hash = ?2
returning *;

-- name: UserUpdatePassword :execrows
update user set password_hash = @password_hash where username = @username;

-- name: UserUpdateRole :exec
update user set role_id = (
  select role.id
  from role
  where role.name = @role
)
where user.username = @username
returning *;

-- name: UserUpdateEnable :exec
update user set enabled = @enabled where username = @username
returning *;

-- name: UserDelete :exec
delete from user where username = @username;
