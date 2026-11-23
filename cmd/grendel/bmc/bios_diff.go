// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	biosDiffTable   *shared.TableFlags
	biosDiffColumns = []shared.Column{
		{Name: "Attribute"},
		{Name: "Value"},
		{Name: "Count"},
		{Name: "Nodes"},
	}
	biosDiffCmd = &cobra.Command{
		Use:   "diff <nodeset | all> [attribute]...",
		Short: "Show the BIOS attributes that differ between nodes",
		Long:  "Show the BIOS attributes that differ between nodes, with the nodes holding each value. A staged change counts as the node's value. Nodes that could not be read are listed after the table and left out of the comparison",
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

			nodes := map[string]map[string]string{}
			var failed failures
			for _, result := range res.Results {
				if !result.Success {
					failed.add(result.Node, result.Message)
					continue
				}
				data := result.Data.Value
				values := map[string]string{}
				for name, raw := range data.Attributes.Value {
					values[name] = rawValue(raw)
				}
				for name, raw := range data.Pending.Value {
					values[name] = rawValue(raw)
				}
				nodes[result.Node] = values
			}

			diff := biosDiff(nodes, args[1:])
			names := make([]string, 0, len(diff))
			for name := range diff {
				names = append(names, name)
			}
			sort.Strings(names)

			t := biosDiffTable.Table()
			for _, name := range names {
				for _, g := range diff[name] {
					t.AppendRow(name, g.Value, strconv.Itoa(len(g.Nodes)), fold(g.Nodes))
				}
			}
			t.Print()
			failed.print()

			fmt.Printf("\n%d attribute(s) differ across %d node(s)\n", len(names), len(nodes))
			return summaryLine(summary(res.Summary))
		},
	}
)

func init() {
	biosDiffTable = shared.RegisterTableFlags(biosDiffCmd, biosDiffColumns)

	biosCmd.AddCommand(biosDiffCmd)
}
