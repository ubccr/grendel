// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package bmc

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/internal/bmc"
)

var (
	jobCmd = &cobra.Command{
		Use:   "job",
		Short: "BMC job commands",
		Long:  `BMC job commands`,
	}

	jobStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "List the status of all jobs",
		Long:  `List the status of all jobs`,
		RunE: func(command *cobra.Command, args []string) error {
			return cmdJobStatus()
		},
	}

	jobClearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Clear all jobs",
		Long:  `Clear all jobs`,
		RunE: func(command *cobra.Command, args []string) error {
			return cmdJobClear()
		},
	}
)

func init() {
	bmcCmd.AddCommand(jobCmd)
	jobCmd.AddCommand(jobStatusCmd)
	jobCmd.AddCommand(jobClearCmd)
}

func cmdJobStatus() error {
	j := bmc.NewJob()
	hostJobs, err := j.GetJobs(hostList)
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Host", "Start Time", "Job Name", "State", "Progress", "Messages"})
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

	for _, hostJob := range hostJobs {
		if len(hostJob.Jobs) == 0 {
			t.AppendRow(table.Row{
				hostJob.Host,
			})
		}
		for _, job := range hostJob.Jobs {
			messages := []string{}
			for _, msg := range job.Messages {
				messages = append(messages, msg.Message)
			}
			layout := "2006-01-02T15:04:05-07:00"
			startTime, err := time.Parse(layout, job.StartTime)
			if err != nil {
				return err
			}
			t.AppendRow(table.Row{
				hostJob.Host,
				startTime.Format(time.DateTime),
				job.Name,
				job.JobState,
				fmt.Sprintf("%d%%", job.PercentComplete),
				strings.Join(messages, ", "),
			}, table.RowConfig{AutoMerge: true})
		}
		t.AppendSeparator()
	}
	t.SetStyle(table.StyleLight)
	t.Render()

	return nil
}

func cmdJobClear() error {
	j := bmc.NewJob()
	jobMessages, err := j.ClearJobs(hostList)
	if err != nil {
		return err
	}

	for _, jobMessage := range jobMessages {
		fmt.Printf("%s\t %s\t %s\n", jobMessage.Host, jobMessage.Status, jobMessage.Msg)
	}

	return nil
}
