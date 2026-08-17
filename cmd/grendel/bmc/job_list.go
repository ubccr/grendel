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
		{Name: "Job Name"},
		{Name: "State"},
		{Name: "Progress"},
		{Name: "Messages"},
	}
	jobListCmd = &cobra.Command{
		Use:   "list <nodeset | all>",
		Short: "List all redfish jobs on the BMC",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			params := client.GETV1BmcJobsParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			nodes, err := apiclient.API.GETV1BmcJobs(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(jobListColumns, jobListFilter).Options(*jobListOptions).Empty("-")

			for _, node := range nodes {
				if len(node.Jobs.Value) == 0 {
					t.AppendRow(
						node.Name.Value,
						"",
						"",
						"",
						"",
					)
					continue
				}

				for _, job := range node.Jobs.Value {
					messages := []string{}
					for _, msg := range job.Value.Messages {
						messages = append(messages, msg.Message.Value)
					}
					t.AppendRow(
						node.Name.Value,
						job.Value.Name.Value,
						job.Value.JobState.Value,
						fmt.Sprintf("%d%%", job.Value.PercentComplete.Value),
						strings.Join(messages, ", "),
					)
				}
			}

			t.Print()
			return nil
		},
	}
)

func init() {
	jobListOptions = shared.RegisterTableFlags(jobListCmd)
	jobListCmd.Flags().StringSliceVar(&jobListFilter, "filter", nil, "filter table columns by name")
	jobListCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(jobListColumns))

	jobCmd.AddCommand(jobListCmd)
}
