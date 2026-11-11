// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	jobListFilter  []string
	jobListOptions *shared.TableOptions
	jobListColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Message"},
		{Name: "Job ID"},
		{Name: "Job Name"},
		{Name: "Job State"},
		{Name: "Progress"},
		{Name: "Job Message"},
	}
	jobListCmd = &cobra.Command{
		Use:     "list <nodeset | all>",
		Aliases: []string{"ls"},
		Short:   "List all redfish jobs on the BMC",
		Args:    cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset, all := selector(args[0])
			params := client.GETV1BmcJobsParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.GETV1BmcJobs(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(jobListColumns, jobListFilter).Options(*jobListOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				state := colorState(result.Success.Value)

				// a node that was reached and has nothing queued still gets a row, which is how an empty queue is told apart from a bmc that never answered
				jobs := result.Data.Value.Jobs.Value
				if len(jobs) == 0 {
					t.AppendRow(result.Node.Value, state, result.Message.Value, "", "", "", "", "")
					continue
				}

				for _, j := range jobs {
					job := j.Value
					t.AppendRow(
						result.Node.Value,
						state,
						result.Message.Value,
						job.ID.Value,
						job.Name.Value,
						job.State.Value,
						fmt.Sprintf("%d%%", job.PercentComplete.Value),
						job.Message.Value,
					)
				}
			}
			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

func init() {
	jobListOptions = shared.RegisterTableFlags(jobListCmd)
	jobListCmd.Flags().StringSliceVar(&jobListFilter, "filter", nil, "filter table columns by name")
	jobListCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(jobListColumns))

	jobCmd.AddCommand(jobListCmd)
}
