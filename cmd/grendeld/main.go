// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	_ "github.com/ubccr/grendel/cmd/grendeld/discover"
	_ "github.com/ubccr/grendel/cmd/grendeld/serve"
	"github.com/ubccr/grendel/cmd/shared"
)

func main() {
	shared.Execute("grendeld", "Bare Metal Provisioning for HPC")
}
