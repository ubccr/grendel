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
	jobClearFilter  []string
	jobClearOptions *shared.TableOptions
	jobClearColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Message"},
	}
	jobClearCmd = &cobra.Command{
		Use:   "clear <nodeset | all> [JID]...",
		Short: "Clear all jobs or by JID",
		Long:  `Clear all jobs or by JID. Clears the whole queue when no JID is given`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			req := &client.BmcClearJobsBody{}
			if jids := args[1:]; len(jids) > 0 {
				ids := make([]client.NilString, 0, len(jids))
				for _, jid := range jids {
					ids = append(ids, client.NewNilString(jid))
				}
				req.Ids = client.NewOptNilNilStringArray(ids)
			} else {
				req.All = client.NewOptBool(true)
			}

			nodeset, all := selector(args[0])
			params := client.POSTV1BmcJobsClearParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcJobsClear(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(jobClearColumns, jobClearFilter).Options(*jobClearOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				t.AppendRow(
					result.Node.Value,
					colorState(result.Success.Value),
					result.Message.Value,
				)
			}
			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

func init() {
	jobClearOptions = shared.RegisterTableFlags(jobClearCmd)
	jobClearCmd.Flags().StringSliceVar(&jobClearFilter, "filter", nil, "filter table columns by name")
	jobClearCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(jobClearColumns))

	jobCmd.AddCommand(jobClearCmd)
}
