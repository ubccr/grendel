// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	selListFilter  []string
	selListLimit   int
	selListOptions *shared.TableOptions
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
			nodeset, all := selector(args[0])
			params := client.GETV1BmcSelParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
				Limit:   client.NewOptString(strconv.Itoa(selListLimit)),
			}

			res, err := apiclient.API.GETV1BmcSel(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(selListColumns, selListFilter).Options(*selListOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				state := colorState(result.Success.Value)

				// a node that was reached with an empty log still gets a row, which is how an empty log is told apart from a bmc that never answered
				entries := result.Data.Value.Entries.Value
				if len(entries) == 0 {
					t.AppendRow(result.Node.Value, state, "", "", "", result.Message.Value, "", "", "")
					continue
				}

				for _, e := range entries {
					entry := e.Value
					t.AppendRow(
						result.Node.Value,
						state,
						entry.ID.Value,
						entry.Created.Value,
						colorSeverity(entry.Severity.Value),
						entry.Message.Value,
						entry.MessageID.Value,
						entry.SensorType.Value,
						entry.Resolution.Value,
					)
				}
			}
			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

// colorSeverity picks out the entries worth reading first, which is the point of looking at a log that runs to thousands of lines
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
	selListOptions = shared.RegisterTableFlags(selListCmd)
	selListCmd.Flags().StringSliceVar(&selListFilter, "filter", []string{"Message ID", "Sensor", "Resolution"}, "filter table columns by name")
	selListCmd.Flags().IntVar(&selListLimit, "limit", 50, "show only this many entries per node, the ones the BMC lists first, which is the newest on most BMCs. 0 reads the whole log and sorts it newest first")

	selListCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(selListColumns))

	selCmd.AddCommand(selListCmd)
}
