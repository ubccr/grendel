// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: kernel.sql

package db

import (
	"context"
	"strings"

	null "github.com/guregu/null/v5"
	"github.com/segmentio/ksuid"
)

const initrdUpsert = `-- name: InitrdUpsert :one
insert into initrd (kernel_id, path)
values (?1, ?2)
on conflict (path, kernel_id)
do update set path = ?2
returning id, kernel_id, path, created_at, updated_at
`

type InitrdUpsertParams struct {
	KernelID int64  `json:"kernel_id"`
	Path     string `json:"path"`
}

func (q *Queries) InitrdUpsert(ctx context.Context, db DBTX, arg InitrdUpsertParams) (Initrd, error) {
	row := db.QueryRowContext(ctx, initrdUpsert, arg.KernelID, arg.Path)
	var i Initrd
	err := row.Scan(
		&i.ID,
		&i.KernelID,
		&i.Path,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const initrdUpsertDelete = `-- name: InitrdUpsertDelete :exec
delete from initrd where kernel_id = ?1 and id not in (/*SLICE:ids*/?)
`

type InitrdUpsertDeleteParams struct {
	KernelID int64   `json:"kernel_id"`
	Ids      []int64 `json:"ids"`
}

func (q *Queries) InitrdUpsertDelete(ctx context.Context, db DBTX, arg InitrdUpsertDeleteParams) error {
	query := initrdUpsertDelete
	var queryParams []interface{}
	queryParams = append(queryParams, arg.KernelID)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	_, err := db.ExecContext(ctx, query, queryParams...)
	return err
}

const kernelAll = `-- name: KernelAll :many
select id, name, image_json from kernel_view
`

func (q *Queries) KernelAll(ctx context.Context, db DBTX) ([]KernelView, error) {
	rows, err := db.QueryContext(ctx, kernelAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []KernelView
	for rows.Next() {
		var i KernelView
		if err := rows.Scan(&i.ID, &i.Name, &i.Image); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const kernelDelete = `-- name: KernelDelete :exec
delete from kernel where name in (/*SLICE:name*/?)
`

func (q *Queries) KernelDelete(ctx context.Context, db DBTX, name []string) error {
	query := kernelDelete
	var queryParams []interface{}
	if len(name) > 0 {
		for _, v := range name {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:name*/?", strings.Repeat(",?", len(name))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:name*/?", "NULL", 1)
	}
	_, err := db.ExecContext(ctx, query, queryParams...)
	return err
}

const kernelFetch = `-- name: KernelFetch :one
/*
 * SPDX-FileCopyrightText: (C) 2019 Grendel Authors
 *
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

select id, name, image_json from kernel_view where name = ?1
`

func (q *Queries) KernelFetch(ctx context.Context, db DBTX, name string) (KernelView, error) {
	row := db.QueryRowContext(ctx, kernelFetch, name)
	var i KernelView
	err := row.Scan(&i.ID, &i.Name, &i.Image)
	return i, err
}

const kernelTemplateUpsert = `-- name: KernelTemplateUpsert :exec
insert into kernel_template (kernel_id, template_id)
values (?1, ?2)
on conflict (kernel_id, template_id)
do nothing
`

type KernelTemplateUpsertParams struct {
	KernelID   int64 `json:"kernel_id"`
	TemplateID int64 `json:"template_id"`
}

func (q *Queries) KernelTemplateUpsert(ctx context.Context, db DBTX, arg KernelTemplateUpsertParams) error {
	_, err := db.ExecContext(ctx, kernelTemplateUpsert, arg.KernelID, arg.TemplateID)
	return err
}

const kernelTemplateUpsertDelete = `-- name: KernelTemplateUpsertDelete :exec
delete from kernel_template where kernel_id = ?1 and template_id not in (/*SLICE:ids*/?)
`

type KernelTemplateUpsertDeleteParams struct {
	KernelID int64   `json:"kernel_id"`
	Ids      []int64 `json:"ids"`
}

func (q *Queries) KernelTemplateUpsertDelete(ctx context.Context, db DBTX, arg KernelTemplateUpsertDeleteParams) error {
	query := kernelTemplateUpsertDelete
	var queryParams []interface{}
	queryParams = append(queryParams, arg.KernelID)
	if len(arg.Ids) > 0 {
		for _, v := range arg.Ids {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:ids*/?", strings.Repeat(",?", len(arg.Ids))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:ids*/?", "NULL", 1)
	}
	_, err := db.ExecContext(ctx, query, queryParams...)
	return err
}

const kernelUpsert = `-- name: KernelUpsert :one
insert into kernel (id, uid, name, version, path, arch_id, command_line, verify)
values (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8)
on conflict (id)
do update set uid = ?2, name = ?3, version = ?4, path = ?5, arch_id = ?6, command_line = ?7, verify = ?8
returning id, uid, name, version, path, arch_id, command_line, verify, created_at, updated_at
`

type KernelUpsertParams struct {
	ID          null.Int64  `json:"id"`
	UID         ksuid.KSUID `json:"uid"`
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Path        string      `json:"path"`
	ArchID      null.Int64  `json:"arch_id"`
	CommandLine null.String `json:"command_line"`
	Verify      bool        `json:"verify"`
}

func (q *Queries) KernelUpsert(ctx context.Context, db DBTX, arg KernelUpsertParams) (Kernel, error) {
	row := db.QueryRowContext(ctx, kernelUpsert,
		arg.ID,
		arg.UID,
		arg.Name,
		arg.Version,
		arg.Path,
		arg.ArchID,
		arg.CommandLine,
		arg.Verify,
	)
	var i Kernel
	err := row.Scan(
		&i.ID,
		&i.UID,
		&i.Name,
		&i.Version,
		&i.Path,
		&i.ArchID,
		&i.CommandLine,
		&i.Verify,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const templateTypeUpsert = `-- name: TemplateTypeUpsert :one
insert into template_type (name, uri_name)
values (?1, ?2)
on conflict (name)
do update set uri_name = ?2
returning id, name, uri_name
`

type TemplateTypeUpsertParams struct {
	Name    string `json:"name"`
	UriName string `json:"uri_name"`
}

func (q *Queries) TemplateTypeUpsert(ctx context.Context, db DBTX, arg TemplateTypeUpsertParams) (TemplateType, error) {
	row := db.QueryRowContext(ctx, templateTypeUpsert, arg.Name, arg.UriName)
	var i TemplateType
	err := row.Scan(&i.ID, &i.Name, &i.UriName)
	return i, err
}

const templateUpsert = `-- name: TemplateUpsert :one
insert into template (name, template_type_id)
values (?1, ?2)
on conflict (name)
do update set name = ?1
returning id, template_type_id, name, created_at, updated_at
`

type TemplateUpsertParams struct {
	Name           string `json:"name"`
	TemplateTypeID int64  `json:"template_type_id"`
}

func (q *Queries) TemplateUpsert(ctx context.Context, db DBTX, arg TemplateUpsertParams) (Template, error) {
	row := db.QueryRowContext(ctx, templateUpsert, arg.Name, arg.TemplateTypeID)
	var i Template
	err := row.Scan(
		&i.ID,
		&i.TemplateTypeID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
