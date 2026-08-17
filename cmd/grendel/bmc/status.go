// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	statusFilter  []string
	statusOptions *shared.TableOptions
	statusColumns = []shared.Column{
		{Name: "Node"},
		{Name: "Host Name"},
		{Name: "Power Status"},
		{Name: "Serial"},
		{Name: "BIOS Version"},
		{Name: "Manufacturer"},
		{Name: "Model"},
		{Name: "Health"},
		{Name: "CPU"},
		{Name: "Memory"},
		{Name: "Boot Order"},
		{Name: "Boot Next"},
		{Name: "OEM"},
	}

	statusCmd = &cobra.Command{
		Use:   "status <nodeset | all>",
		Short: "Check BMC status",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			params := client.GETV1BmcParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := apiclient.API.GETV1Bmc(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(statusColumns, statusFilter).Options(*statusOptions).Empty("-")
			for _, o := range res {
				bootOrder := make([]string, 0, len(o.BootOrder.Value))
				for _, s := range o.BootOrder.Value {
					bootOrder = append(bootOrder, s.Value)
				}

				oem, err := json.MarshalIndent(o.OemDell, "", "  ")
				if err != nil {
					log.WithField("node", o.Name.Value).Error(fmt.Errorf("failed to marshal OEM json: %w", err))
				}

				t.AppendRow(
					o.Name.Value,
					o.HostName.Value,
					o.PowerStatus.Value,
					o.SerialNumber.Value,
					o.BiosVersion.Value,
					o.Manufacturer.Value,
					o.Model.Value,
					o.Health.Value,
					fmt.Sprintf("%d", o.ProcessorCount.Value),
					fmt.Sprintf("%.2f", o.TotalMemory.Value),
					strings.Join(bootOrder, ","),
					o.BootNext.Value,
					string(oem),
				)
			}

			t.Print()
			return nil
		},
	}
)

func init() {
	statusOptions = shared.RegisterTableFlags(statusCmd)
	statusCmd.Flags().StringSliceVar(&statusFilter, "filter", []string{"Host Name", "Manufacturer", "CPU", "Memory", "Boot Order", "Boot Next", "OEM"}, "filter table columns by name")

	statusCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(statusColumns))

	bmcCmd.AddCommand(statusCmd)
}
