// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

// Package version holds the build version on its own, so that a binary can report it without linking the api server and the web ui embedded alongside it
package version

// Version of Grendel, set at build time with -X github.com/ubccr/grendel/internal/version.Version
var Version = "vDEV"
