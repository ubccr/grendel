// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

//go:build !unix

package storetest

import "time"

// cpuTime has no portable implementation off unix; benchmarks simply omit the
// CPU metrics there.
func cpuTime() (user, sys time.Duration, ok bool) {
	return 0, 0, false
}
