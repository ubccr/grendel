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
	secureBootStatusFilter  []string
	secureBootStatusOptions *shared.TableOptions
	secureBootStatusColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Enabled"},
		{Name: "Pending"},
		{Name: "Mode"},
		{Name: "Current Boot"},
		{Name: "Databases"},
		{Name: "Message"},
	}
	secureBootStatusCmd = &cobra.Command{
		Use:   "status <nodeset | all>",
		Short: "Show whether UEFI secure boot is on",
		Long:  "Show whether UEFI secure boot is on. Enabled is what the next boot uses, Pending is a change staged for it, and Current Boot is whether secure boot was in effect for the boot the node is running now",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset, all := selector(args[0])
			params := client.GETV1BmcSecurebootParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.GETV1BmcSecureboot(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(secureBootStatusColumns, secureBootStatusFilter).Options(*secureBootStatusOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				if !result.Success.Value {
					t.AppendRow(result.Node.Value, colorState(false), "", "", "", "", "", result.Message.Value)
					continue
				}

				data := result.Data.Value
				pending := ""
				if p, ok := data.PendingEnabled.Get(); ok {
					pending = strconv.FormatBool(p)
				}
				t.AppendRow(
					result.Node.Value,
					colorState(true),
					colorEnabled(data.Enabled.Value),
					pending,
					data.Mode.Value,
					data.CurrentBoot.Value,
					strings.Join(data.Databases, ","),
					result.Message.Value,
				)
			}
			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

func colorEnabled(on bool) string {
	if on {
		return text.FgHiGreen.Sprint("true")
	}
	return text.FgHiYellow.Sprint("false")
}

func init() {
	secureBootStatusOptions = shared.RegisterTableFlags(secureBootStatusCmd)
	secureBootStatusCmd.Flags().StringSliceVar(&secureBootStatusFilter, "filter", []string{"Databases", "Message"}, "filter table columns by name")
	secureBootStatusCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(secureBootStatusColumns))

	secureBootCmd.AddCommand(secureBootStatusCmd)
}
