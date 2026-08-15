// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendelctl/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	jobClearFilter  []string
	jobClearOptions *shared.TableOptions
	jobClearColumns = []shared.Column{
		{Name: "Node"},
		{Name: "Status"},
		{Name: "Message"},
	}
	jobClearCmd = &cobra.Command{
		Use:   "clear <nodeset | all> [JID]...",
		Short: "Clear all jobs or by JID",
		Long:  `Clear all jobs or by JID. Defaults to JID_CLEARALL to clear all jobs`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			jids := args[1:]
			if len(jids) == 0 {
				jids = []string{"JID_CLEARALL"}
			}
			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			params := client.DELETEV1BmcJobsJidsParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
				Jids:    strings.Join(jids, ","),
			}
			jobs, err := apiclient.API.DELETEV1BmcJobsJids(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(jobClearColumns, jobClearFilter).Options(*jobClearOptions).Empty("-")

			for _, job := range jobs {
				t.AppendRow(
					job.Host.Value,
					job.Status.Value,
					job.Msg.Value,
				)
			}

			t.Print()
			return nil
		},
	}
)

func init() {
	jobClearOptions = shared.RegisterTableFlags(jobClearCmd)
	jobClearCmd.Flags().StringSliceVar(&jobClearFilter, "filter", nil, "filter table columns by name")
	jobClearCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(jobClearColumns))

	jobCmd.AddCommand(jobClearCmd)
}
