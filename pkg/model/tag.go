// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

type TagList []Tag

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TagCount is a tag as it is written on a node, key or key=value, and how many nodes have it
type TagCount struct {
	Tag   string `json:"tag"`
	Nodes int64  `json:"nodes"`
}

type TagCountList []TagCount
