// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

type DataDump struct {
	Users  []User
	Hosts  HostList
	Images BootImageList
}
