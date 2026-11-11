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
	biosDiffFilter  []string
	biosDiffOptions *shared.TableOptions
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
			nodeset, all := selector(args[0])
			params := client.GETV1BmcBiosParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			if len(args) > 1 {
				params.Names = client.NewOptString(strings.Join(args[1:], ","))
			}

			res, err := apiclient.API.GETV1BmcBios(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			nodes := map[string]map[string]string{}
			for _, r := range res.Results.Value {
				result := r.Value
				if !result.Success.Value {
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
				nodes[result.Node.Value] = values
			}

			diff := biosDiff(nodes, args[1:])
			names := make([]string, 0, len(diff))
			for name := range diff {
				names = append(names, name)
			}
			sort.Strings(names)

			t := shared.NewTable(biosDiffColumns, biosDiffFilter).Options(*biosDiffOptions).Empty("-")
			for _, name := range names {
				for _, g := range diff[name] {
					t.AppendRow(name, g.Value, strconv.Itoa(len(g.Nodes)), fold(g.Nodes))
				}
			}
			t.Print()

			fmt.Printf("\n%d attribute(s) differ across %d node(s)\n", len(names), len(nodes))
			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

func init() {
	biosDiffOptions = shared.RegisterTableFlags(biosDiffCmd)
	biosDiffCmd.Flags().StringSliceVar(&biosDiffFilter, "filter", nil, "filter table columns by name")
	biosDiffCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(biosDiffColumns))

	biosCmd.AddCommand(biosDiffCmd)
}
