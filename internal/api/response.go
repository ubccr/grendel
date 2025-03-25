// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

type GenericResponse struct {
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	Changed int    `json:"changed"`
}
