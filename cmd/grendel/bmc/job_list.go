// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	jobListTable   *shared.TableFlags
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
			params := client.GETV1BmcJobsParams(selected(args[0]))

			res, err := apiclient.API.GETV1BmcJobs(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := jobListTable.Table()
			for _, result := range res.Results {
				state := colorState(result.Success)

				// a node that was reached and has nothing queued still gets a row, which is how an empty queue is told apart from a bmc that never answered
				jobs := result.Data.Value.Jobs.Value
				if len(jobs) == 0 {
					t.AppendRow(result.Node, state, result.Message, "", "", "", "", "")
					continue
				}

				for _, job := range jobs {
					t.AppendRow(
						result.Node,
						state,
						result.Message,
						job.ID,
						job.Name,
						job.State,
						fmt.Sprintf("%d%%", job.PercentComplete),
						job.Message,
					)
				}
			}
			t.Print()

			return summaryLine(summary(res.Summary))
		},
	}
)

func init() {
	jobListTable = shared.RegisterTableFlags(jobListCmd, jobListColumns)

	jobCmd.AddCommand(jobListCmd)
}
