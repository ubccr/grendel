// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	_ "github.com/ubccr/grendel/cmd/grendel/auth"
	_ "github.com/ubccr/grendel/cmd/grendel/bmc"
	_ "github.com/ubccr/grendel/cmd/grendel/db"
	_ "github.com/ubccr/grendel/cmd/grendel/image"
	_ "github.com/ubccr/grendel/cmd/grendel/lldp"
	_ "github.com/ubccr/grendel/cmd/grendel/node"
	_ "github.com/ubccr/grendel/cmd/grendel/status"
	_ "github.com/ubccr/grendel/cmd/grendel/version"
	"github.com/ubccr/grendel/cmd/shared"
)

func main() {
	apiclient.Register()
	shared.Execute("grendel", "Grendel command line client")
}
