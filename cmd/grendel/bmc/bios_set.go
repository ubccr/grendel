// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	biosSetFilter    []string
	biosSetApplyTime string
	biosSetOptions   *shared.TableOptions
	biosSetColumns   = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Message"},
	}
	biosSetCmd = &cobra.Command{
		Use:   "set <nodeset | all> <Attribute=Value>...",
		Short: "Stage BIOS attribute changes",
		Long:  "Stage BIOS attribute changes. Values are converted to the type each attribute holds, and a node is left unchanged if any value does not fit it. Use 'bmc bios describe' to list an attribute's allowed values",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			attrs, err := parseAssignments(args[1:])
			if err != nil {
				return err
			}

			req := &client.BmcSetBiosBody{Attributes: client.NewOptBmcSetBiosBodyAttributes(attrs)}
			if biosSetApplyTime != "" {
				req.ApplyTime = client.NewOptString(biosSetApplyTime)
			}

			nodeset, all := selector(args[0])
			params := client.POSTV1BmcBiosParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcBios(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(biosSetColumns, biosSetFilter).Options(*biosSetOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				t.AppendRow(result.Node.Value, colorState(result.Success.Value), result.Message.Value)
			}
			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

func init() {
	biosSetOptions = shared.RegisterTableFlags(biosSetCmd)
	biosSetCmd.Flags().StringSliceVar(&biosSetFilter, "filter", nil, "filter table columns by name")
	biosSetCmd.Flags().StringVar(&biosSetApplyTime, "apply-time", "", "when the BMC applies the change: Immediate, OnReset, AtMaintenanceWindowStart or InMaintenanceWindowOnReset. Defaults to the BMC's own default. Immediate can reboot the host")
	biosSetCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(biosSetColumns))
	biosSetCmd.RegisterFlagCompletionFunc("apply-time", cobra.FixedCompletions([]string{"Immediate", "OnReset", "AtMaintenanceWindowStart", "InMaintenanceWindowOnReset"}, cobra.ShellCompDirectiveNoFileComp))

	biosCmd.AddCommand(biosSetCmd)
}
