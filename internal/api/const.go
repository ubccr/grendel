// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

const (
	DefaultPort = 8080

	ContextKeyUsername GrendelAuthContext = "username"
	ContextKeyRole     GrendelAuthContext = "role"
)

type GrendelAuthContext string
