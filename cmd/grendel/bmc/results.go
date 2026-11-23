// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"fmt"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

// selection is the node selector every bmc route takes. The generated params of a route that takes nothing else have these fields exactly, so they convert from it
type selection struct {
	Nodeset client.OptString
	Tags    client.OptString
	All     client.OptBool
	Accept  client.OptString
}

// selected maps the positional nodeset argument and --tags onto a selection. Every node is asked for explicitly, because an empty nodeset selects nothing
func selected(arg string) selection {
	sel := selection{Tags: client.NewOptString(strings.Join(tags, ","))}
	if arg == "all" {
		sel.All = client.NewOptBool(true)
	} else {
		sel.Nodeset = client.NewOptString(arg)
	}
	return sel
}

// target names the nodes a command is about to change, for a confirmation prompt
func target(arg string) string {
	if len(tags) == 0 {
		return arg
	}
	return fmt.Sprintf("%s with tags %s", arg, strings.Join(tags, ","))
}

// colorState makes the rows that need attention findable in a table of several hundred
func colorState(success bool) string {
	if success {
		return text.FgHiGreen.Sprint("success")
	}
	return text.FgHiRed.Sprint("failed")
}

// summary is how a run went. The generated clients have one summary type per route, all with these fields, so they convert to it
type summary struct {
	DurationMs  int64
	Failed      int
	FailedNodes string
	Succeeded   int
	Total       int
}

// summaryLine reports the run and returns the error that makes the command exit non-zero. A bmc route answers 200 for a completed run even when every node in it failed, so the exit code is what a script has to go on
func summaryLine(s summary) error {
	fmt.Printf("\n%d node(s): %d succeeded, %d failed in %s\n",
		s.Total, s.Succeeded, s.Failed,
		(time.Duration(s.DurationMs) * time.Millisecond).Round(time.Millisecond))

	if s.Failed > 0 {
		return fmt.Errorf("%d of %d node(s) failed: %s", s.Failed, s.Total, s.FailedNodes)
	}

	return nil
}

// failures are the nodes of a run that failed and why, for a table whose rows are a node's items rather than the node
type failures []string

func (f *failures) add(node, message string) {
	*f = append(*f, fmt.Sprintf("%s: %s", node, message))
}

func (f failures) print() {
	if len(f) == 0 {
		return
	}
	fmt.Println("\nFailed:")
	for _, line := range f {
		fmt.Println("  " + text.FgHiRed.Sprint(line))
	}
}

var ackColumns = []shared.Column{
	{Name: "Node"},
	{Name: "State"},
	{Name: "Message"},
}

// printAcks prints how a change went on each node and returns the error that sets the exit code
func printAcks(res *client.BmcResponseModelBmcAck, table *shared.TableFlags) error {
	t := table.Table()
	for _, result := range res.Results {
		t.AppendRow(result.Node, colorState(result.Success), result.Message)
	}
	t.Print()

	return summaryLine(summary(res.Summary))
}

// ackCommand gives a command that changes something the table flags, and has it print what run returns. A run that returns no response chose not to act, as when a confirmation is declined
func ackCommand(command *cobra.Command, run func(command *cobra.Command, args []string) (*client.BmcResponseModelBmcAck, error)) *cobra.Command {
	table := shared.RegisterTableFlags(command, ackColumns)
	command.RunE = func(command *cobra.Command, args []string) error {
		res, err := run(command, args)
		if err != nil {
			return apiclient.NewApiError(err)
		}
		if res == nil {
			return nil
		}
		return printAcks(res, table)
	}
	return command
}
