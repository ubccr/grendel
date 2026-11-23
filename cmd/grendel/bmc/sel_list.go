// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	selListLimit   int
	selListTable   *shared.TableFlags
	selListColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Entry"},
		{Name: "Created"},
		{Name: "Severity"},
		{Name: "Message"},
		{Name: "Message ID"},
		{Name: "Sensor"},
		{Name: "Resolution"},
	}
	selListCmd = &cobra.Command{
		Use:     "list <nodeset | all>",
		Aliases: []string{"ls"},
		Short:   "List the system event log on the BMC",
		Args:    cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			sel := selected(args[0])
			params := client.GETV1BmcSelParams{Nodeset: sel.Nodeset, Tags: sel.Tags, All: sel.All, Limit: client.NewOptInt(selListLimit)}

			res, err := apiclient.API.GETV1BmcSel(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := selListTable.Table()
			for _, result := range res.Results {
				state := colorState(result.Success)

				// a node that was reached with an empty log still gets a row, which is how an empty log is told apart from a bmc that never answered
				entries := result.Data.Value.Entries.Value
				if len(entries) == 0 {
					t.AppendRow(result.Node, state, "", "", "", result.Message, "", "", "")
					continue
				}

				for _, entry := range entries {
					t.AppendRow(
						result.Node,
						state,
						entry.ID,
						entry.Created,
						colorSeverity(entry.Severity),
						entry.Message,
						entry.MessageID,
						entry.SensorType,
						entry.Resolution,
					)
				}
			}
			t.Print()

			return summaryLine(summary(res.Summary))
		},
	}
)

// colorSeverity picks out the entries worth reading first in a log that runs to thousands of lines
func colorSeverity(severity string) string {
	switch severity {
	case "OK":
		return text.FgHiGreen.Sprint(severity)
	case "Warning":
		return text.FgHiYellow.Sprint(severity)
	case "Critical":
		return text.FgHiRed.Sprint(severity)
	default:
		return severity
	}
}

func init() {
	selListTable = shared.RegisterTableFlags(selListCmd, selListColumns, "Message ID", "Sensor", "Resolution")
	selListCmd.Flags().IntVar(&selListLimit, "limit", 50, "show only this many entries per node, the ones the BMC lists first, which is the newest on most BMCs. 0 reads the whole log and sorts it newest first")

	selCmd.AddCommand(selListCmd)
}
