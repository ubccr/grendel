// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

//go:generate go run ./gen ../../configs/grendel.toml.sample

const sampleHeader = `# SPDX-FileCopyrightText: (C) 2019 Grendel Authors
#
# SPDX-License-Identifier: GPL-3.0-or-later

`

// SampleFile is the content of configs/grendel.toml.sample
func SampleFile() string { return sampleFile() }

func sampleFile() string { return sampleHeader + Sample() }
