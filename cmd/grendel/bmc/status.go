// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"encoding/json"
	"fmt"
	"strconv"
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
		{Name: "State"},
		{Name: "Message"},
		{Name: "Host Name"},
		{Name: "Power Status"},
		{Name: "Serial"},
		{Name: "BIOS Version"},
		{Name: "Manufacturer"},
		{Name: "Model"},
		{Name: "Health"},
		{Name: "Locate LED"},
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
			nodeset, all := selector(args[0])
			params := client.GETV1BmcStatusParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.GETV1BmcStatus(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(statusColumns, statusFilter).Options(*statusOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				// data is null unless the node was reached, so every other column stays empty and the row still names the node and why it failed
				o := result.Data.Value

				bootOrder := make([]string, 0, len(o.BootOrder.Value))
				for _, s := range o.BootOrder.Value {
					bootOrder = append(bootOrder, s.Value)
				}

				var oem []byte
				if o.OemDell.Set && !o.OemDell.Null {
					oem, err = json.MarshalIndent(o.OemDell, "", "  ")
					if err != nil {
						log.WithField("node", result.Node.Value).Error(fmt.Errorf("failed to marshal OEM json: %w", err))
					}
				}

				t.AppendRow(
					result.Node.Value,
					colorState(result.Success.Value),
					result.Message.Value,
					o.HostName.Value,
					o.PowerStatus.Value,
					o.SerialNumber.Value,
					o.BiosVersion.Value,
					o.Manufacturer.Value,
					o.Model.Value,
					o.Health.Value,
					o.LocationIndicator.Value,
					optInt(o.ProcessorCount),
					optFloat(o.TotalMemory),
					strings.Join(bootOrder, "\n"),
					o.BootNext.Value,
					string(oem),
				)
			}

			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

// optInt and optFloat leave the cell empty when the field was absent, so an unreached node shows the table's placeholder rather than a made up zero
func optInt(v client.OptInt) string {
	if !v.Set {
		return ""
	}
	return strconv.Itoa(v.Value)
}

func optFloat(v client.OptFloat32) string {
	if !v.Set {
		return ""
	}
	return fmt.Sprintf("%.2f", v.Value)
}

func init() {
	statusOptions = shared.RegisterTableFlags(statusCmd)
	statusCmd.Flags().StringSliceVar(&statusFilter, "filter", []string{"Host Name", "Manufacturer", "CPU", "Memory", "Boot Order", "Boot Next", "OEM"}, "filter table columns by name")

	statusCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(statusColumns))

	bmcCmd.AddCommand(statusCmd)
}
