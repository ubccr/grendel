// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	biosGetPending bool
	biosGetTable   *shared.TableFlags
	biosGetColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Attribute"},
		{Name: "Value"},
		{Name: "Pending"},
		{Name: "Message"},
	}
	biosGetCmd = &cobra.Command{
		Use:   "get <nodeset | all> [attribute]...",
		Short: "Show BIOS attributes and any staged changes",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			sel := selected(args[0])
			params := client.GETV1BmcBiosParams{Nodeset: sel.Nodeset, Tags: sel.Tags, All: sel.All}
			if len(args) > 1 {
				params.Names = client.NewOptString(strings.Join(args[1:], ","))
			}

			res, err := apiclient.API.GETV1BmcBios(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := biosGetTable.Table()
			var failed failures
			for _, result := range res.Results {
				state := colorState(result.Success)
				if !result.Success {
					failed.add(result.Node, result.Message)
				}
				data := result.Data.Value
				attrs := data.Attributes.Value
				pending := data.Pending.Value

				names := make([]string, 0, len(attrs))
				for name := range attrs {
					if !biosGetPending || pending[name] != nil {
						names = append(names, name)
					}
				}
				sort.Strings(names)

				// a node with nothing to show still gets a row, so every node of the run is in the table
				if len(names) == 0 {
					t.AppendRow(result.Node, state, "", "", "", result.Message)
					continue
				}
				for _, name := range names {
					t.AppendRow(result.Node, state, name, rawValue(attrs[name]), rawValue(pending[name]), result.Message)
				}
			}
			t.Print()
			failed.print()

			return summaryLine(summary(res.Summary))
		},
	}
)

func init() {
	biosGetTable = shared.RegisterTableFlags(biosGetCmd, biosGetColumns, "Message")
	biosGetCmd.Flags().BoolVar(&biosGetPending, "pending", false, "show only attributes with a staged change")

	biosCmd.AddCommand(biosGetCmd)
}
