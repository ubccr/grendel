/*
 * SPDX-FileCopyrightText: (C) 2019 Grendel Authors
 *
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

-- name: KernelFetch :one
select * from kernel_view where name = @name; 

-- name: KernelAll :many
select * from kernel_view;
;

-- name: KernelUpsert :one
insert into kernel (id, uid, name, version, path, arch_id, command_line, verify)
values (sqlc.narg(id), @uid, @name, @version, @path, @arch_id, @command_line, @verify)
on conflict (id)
do update set uid = ?2, name = ?3, version = ?4, path = ?5, arch_id = ?6, command_line = ?7, verify = ?8
returning *;

-- name: InitrdUpsert :exec
insert into initrd (kernel_id, path)
values (@kernel_id, @path)
on conflict (path, kernel_id)
do nothing
returning *;

-- name: TemplateTypeUpsert :one
insert into template_type (name, uri_name)
values (@name, @uri_name)
on conflict (name)
do update set uri_name = ?2
returning *;

-- name: TemplateUpsert :one
insert into template (name, template_type_id)
values (@name, @template_type_id)
on conflict (name)
do update set name = ?1
returning *;

-- name: KernelTemplateUpsert :exec
insert into kernel_template (kernel_id, template_id)
values (@kernel_id, @template_id)
on conflict (kernel_id, template_id)
do nothing;

-- name: KernelDelete :exec
delete from kernel where name in (sqlc.slice(name));
