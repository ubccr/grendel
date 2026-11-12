// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
	"github.com/ubccr/grendel/pkg/model"
)

var (
	statusFilter  []string
	statusOptions *shared.TableOptions
	statusColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Started"},
		{Name: "Updated"},
		{Name: "Elapsed"},
		{Name: "Image"},
		{Name: "Failed Units"},
		{Name: "Message"},
	}

	statusCmd = &cobra.Command{
		Use:               "status {nodeset | all}",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: nodesetCompletion,
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]
			if nodeset == "all" {
				nodeset = ""
			}
			params := client.GETV1NodesProvisionStatusParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := apiclient.API.GETV1NodesProvisionStatus(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			now := time.Now()
			t := shared.NewTable(statusColumns, statusFilter).Options(*statusOptions).Empty("-")
			for _, s := range res {
				var units []string
				for _, u := range s.FailedUnits.Value {
					units = append(units, u.Name.Value)
				}

				t.AppendRow(
					s.Node.Value,
					colorProvisionState(s.State.Value),
					formatTime(s.StartedAt),
					formatTime(s.UpdatedAt),
					elapsed(s, now),
					imageCell(s),
					strings.Join(units, "\n"),
					s.Message.Value,
				)
			}
			t.Print()
			return nil
		},
	}
)

// imageCell is the image version the node reports running, in red with the version it was sent when they differ
func imageCell(s client.ProvisionStatus) string {
	running := s.Facts.Value[model.FactImageVersion]
	sent := s.BootImageVersion.Value
	if running == "" {
		return ""
	}
	if sent != "" && running != sent {
		return text.FgHiRed.Sprintf("%s (sent %s)", running, sent)
	}
	return running
}

func colorProvisionState(state string) string {
	switch state {
	case "completed":
		return text.FgHiGreen.Sprint(state)
	case "failed":
		return text.FgHiRed.Sprint(state)
	}
	return text.FgHiYellow.Sprint(state)
}

func formatTime(t client.OptDateTime) string {
	if !t.Set || t.Value.IsZero() {
		return ""
	}
	return t.Value.Local().Format(time.DateTime)
}

// elapsed is how long the provision took when it finished, or has taken so far when it hasn't
func elapsed(s client.ProvisionStatus, now time.Time) string {
	if !s.StartedAt.Set {
		return ""
	}
	end := now
	switch {
	case s.CompletedAt.Set && !s.CompletedAt.Null:
		end = s.CompletedAt.Value
	case s.State.Value == "failed" && s.UpdatedAt.Set:
		end = s.UpdatedAt.Value
	}
	return end.Sub(s.StartedAt.Value).Round(time.Second).String()
}

func init() {
	statusOptions = shared.RegisterTableFlags(statusCmd)
	statusCmd.Flags().StringSliceVar(&statusFilter, "filter", nil, "filter table columns by name")
	statusCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(statusColumns))
	nodeCmd.AddCommand(statusCmd)
}
