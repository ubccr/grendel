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
	biosGetFilter  []string
	biosGetPending bool
	biosGetOptions *shared.TableOptions
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

			t := shared.NewTable(biosGetColumns, biosGetFilter).Options(*biosGetOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				state := colorState(result.Success.Value)
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

				// a node with nothing to show still gets a row, so one that failed says why
				if len(names) == 0 {
					t.AppendRow(result.Node.Value, state, "", "", "", result.Message.Value)
					continue
				}
				for _, name := range names {
					t.AppendRow(result.Node.Value, state, name, rawValue(attrs[name]), rawValue(pending[name]), result.Message.Value)
				}
			}
			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

func init() {
	biosGetOptions = shared.RegisterTableFlags(biosGetCmd)
	biosGetCmd.Flags().StringSliceVar(&biosGetFilter, "filter", []string{"Message"}, "filter table columns by name")
	biosGetCmd.Flags().BoolVar(&biosGetPending, "pending", false, "show only attributes with a staged change")
	biosGetCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(biosGetColumns))

	biosCmd.AddCommand(biosGetCmd)
}
