// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Command grendel is the Grendel server. It runs the services and the privileged node discovery commands, everything that talks to the api instead lives in grendelctl
package main

import (
	_ "github.com/ubccr/grendel/cmd/grendel/discover"
	_ "github.com/ubccr/grendel/cmd/grendel/serve"
	"github.com/ubccr/grendel/cmd/shared"
)

func main() {
	shared.Execute("grendel", "Bare Metal Provisioning for HPC")
}
