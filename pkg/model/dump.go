// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

type DataDump struct {
	Users  []User        `json:"Users"`
	Hosts  HostList      `json:"Hosts"`
	Images BootImageList `json:"Images"`
	Roles  RoleViewList  `json:"Roles"`
	// Settings is the configuration stored in the database, without the secrets grendel generated for itself
	Settings []StoredSetting `json:"Settings,omitempty"`
}

type StoredSetting struct {
	Key string `json:"key"`
	// Value is the setting's value as JSON text
	Value string `json:"value"`
}
