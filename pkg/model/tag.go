// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

type TagList []Tag

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
