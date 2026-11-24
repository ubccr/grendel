// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Command gen writes the sample config file from the settings registry
package main

import (
	"log"
	"os"

	"github.com/ubccr/grendel/internal/settings"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: gen <path>")
	}
	if err := os.WriteFile(os.Args[1], []byte(settings.SampleFile()), 0o644); err != nil {
		log.Fatal(err)
	}
}
