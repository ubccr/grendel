// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Command grendelctl is the Grendel command line client. It reaches a running server over the api and links none of the server itself, so it builds without cgo
package main

import (
	"github.com/ubccr/grendel/cmd/grendelctl/apiclient"
	_ "github.com/ubccr/grendel/cmd/grendelctl/auth"
	_ "github.com/ubccr/grendel/cmd/grendelctl/bmc"
	_ "github.com/ubccr/grendel/cmd/grendelctl/db"
	_ "github.com/ubccr/grendel/cmd/grendelctl/image"
	_ "github.com/ubccr/grendel/cmd/grendelctl/lldp"
	_ "github.com/ubccr/grendel/cmd/grendelctl/node"
	_ "github.com/ubccr/grendel/cmd/grendelctl/status"
	_ "github.com/ubccr/grendel/cmd/grendelctl/version"
	"github.com/ubccr/grendel/cmd/shared"
)

func main() {
	apiclient.Register()
	shared.Execute("grendelctl", "Grendel command line client")
}
