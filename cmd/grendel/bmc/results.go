// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"fmt"
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/ubccr/grendel/pkg/client"
)

// selector maps the positional nodeset argument onto the bmc query. Every node is asked for explicitly, because an empty nodeset selects nothing
func selector(arg string) (nodeset, all client.OptString) {
	if arg == "all" {
		return client.OptString{}, client.NewOptString("true")
	}
	return client.NewOptString(arg), client.OptString{}
}

// colorState makes the rows that need attention findable in a table of several hundred
func colorState(success bool) string {
	if success {
		return text.FgHiGreen.Sprint("success")
	}
	return text.FgHiRed.Sprint("failed")
}

// summaryLine reports the run and returns the error that makes the command exit non-zero. A bmc route answers 200 for a completed run even when every node in it failed, so the exit code is what a script has to go on
func summaryLine(total, succeeded, failed int, failedNodes string, durationMs int64) error {
	fmt.Printf("\n%d node(s): %d succeeded, %d failed in %s\n",
		total, succeeded, failed,
		(time.Duration(durationMs) * time.Millisecond).Round(time.Millisecond))

	if failed > 0 {
		return fmt.Errorf("%d of %d node(s) failed: %s", failed, total, failedNodes)
	}

	return nil
}
