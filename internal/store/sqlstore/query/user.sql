/*
 * SPDX-FileCopyrightText: (C) 2019 Grendel Authors
 *
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

-- name: UserCount :one
select count(*) from user;

-- name: UserFetch :one
select * from user where username = @username;

-- name: UserList :many
select * from user;

-- name: UserCreate :one
insert into user (username, password_hash, role) 
values (@username, @password_hash, @role)
on conflict (username)
do update set password_hash = ?2, role = ?3
returning *;

-- name: UserUpdate :exec
update user set role = @role where username = @username
returning *;

-- name: UserDelete :exec
delete from user where username = @username;
