// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	jobCmd = &cobra.Command{
		Use:   "job",
		Short: "BMC job commands",
	}

	jobShowCmd = &cobra.Command{
		Use:   "show",
		Short: "List all redfish jobs on the BMC",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			var err error
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			params := client.GETV1BmcJobsParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := gc.GETV1BmcJobs(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Host", "Job Name", "State", "Progress", "Messages"})
			t.SetColumnConfigs([]table.ColumnConfig{
				{
					Name:      "Host",
					AutoMerge: true,
				},
				{
					Name:     "Messages",
					WidthMax: 64,
				},
			})

			for _, hostJob := range res {
				if len(hostJob.Jobs.Value) == 0 {
					t.AppendRow(table.Row{
						hostJob.Name.Value,
					})
				}
				for _, job := range hostJob.Jobs.Value {
					messages := []string{}
					for _, msg := range job.Value.Messages {
						messages = append(messages, msg.Message.Value)
					}
					t.AppendRow(table.Row{
						hostJob.Name.Value,
						job.Value.Name.Value,
						job.Value.JobState.Value,
						fmt.Sprintf("%d%%", job.Value.PercentComplete.Value),
						strings.Join(messages, ", "),
					}, table.RowConfig{AutoMerge: true})
				}
				t.AppendSeparator()
			}
			t.SetStyle(table.StyleLight)
			t.Render()

			return nil
		},
	}

	jobClearCmd = &cobra.Command{
		Use:   "clear {nodeset | all} [JIDs...]",
		Short: "Clear all jobs or by JID",
		Long:  `Clear all jobs or by JID. Defaults to JID_CLEARALL to clear all jobs`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) == 1 {
				args = append(args, "JID_CLEARALL")
			}
			var err error
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			params := client.DELETEV1BmcJobsJidsParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
				Jids:    strings.Join(args, ","),
			}
			res, err := gc.DELETEV1BmcJobsJids(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			for _, jobMessage := range res {
				fmt.Printf("%s\t %s\t %s\n", jobMessage.Host.Value, jobMessage.Status.Value, jobMessage.Msg.Value)
			}

			return nil
		},
	}
)

func init() {
	bmcCmd.AddCommand(jobCmd)
	jobCmd.AddCommand(jobShowCmd)
	jobCmd.AddCommand(jobClearCmd)
}
