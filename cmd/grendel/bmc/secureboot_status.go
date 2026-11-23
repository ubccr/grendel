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
	secureBootStatusTable   *shared.TableFlags
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
			sel := selected(args[0])
			res, err := apiclient.API.GETV1BmcSecureboot(command.Context(), client.GETV1BmcSecurebootParams{Nodeset: sel.Nodeset, Tags: sel.Tags, All: sel.All})
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := secureBootStatusTable.Table()
			for _, result := range res.Results {
				if !result.Success {
					t.AppendRow(result.Node, colorState(false), "", "", "", "", "", result.Message)
					continue
				}

				data := result.Data.Value
				pending := ""
				if p, ok := data.PendingEnabled.Get(); ok {
					pending = strconv.FormatBool(p)
				}
				t.AppendRow(
					result.Node,
					colorState(true),
					colorEnabled(data.Enabled),
					pending,
					data.Mode,
					data.CurrentBoot,
					strings.Join(data.Databases, ","),
					result.Message,
				)
			}
			t.Print()

			return summaryLine(summary(res.Summary))
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
	secureBootStatusTable = shared.RegisterTableFlags(secureBootStatusCmd, secureBootStatusColumns, "Databases")

	secureBootCmd.AddCommand(secureBootStatusCmd)
}
