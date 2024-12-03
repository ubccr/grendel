// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"github.com/ubccr/grendel/cmd"
	_ "github.com/ubccr/grendel/cmd/all"
)

func main() {
	cmd.Execute()
}
