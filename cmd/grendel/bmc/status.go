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
	statusTable   *shared.TableFlags
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
		{Name: "Asset Tag"},
		{Name: "Location"},
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
			res, err := bmcStatus(command, args[0])
			if err != nil {
				return err
			}

			t := statusTable.Table()
			for _, result := range res.Results {
				// data is null unless the node was reached, so every other column stays empty and the row still names the node and why it failed
				o, reached := result.Data.Get()

				var oem []byte
				if dell, ok := o.OemDell.Get(); ok {
					oem, err = json.MarshalIndent(dell, "", "  ")
					if err != nil {
						log.WithField("node", result.Node).Error(fmt.Errorf("failed to marshal OEM json: %w", err))
					}
				}

				// a count the bmc never gave must not show as a made up zero
				cpu, memory := "", ""
				if reached {
					cpu, memory = strconv.Itoa(o.ProcessorCount), fmt.Sprintf("%.2f", o.TotalMemory)
				}

				t.AppendRow(
					result.Node,
					colorState(result.Success),
					result.Message,
					o.HostName,
					o.PowerStatus,
					o.SerialNumber,
					o.BiosVersion,
					o.Manufacturer,
					o.Model,
					o.Health,
					o.LocationIndicator,
					o.AssetTag,
					formatLocation(location(o.Location.Value)),
					cpu,
					memory,
					strings.Join(o.BootOrder.Value, "\n"),
					o.BootNext,
					string(oem),
				)
			}

			t.Print()

			return summaryLine(summary(res.Summary))
		},
	}
)

func bmcStatus(command *cobra.Command, selection string) (*client.BmcResponseModelRedfishSystem, error) {
	params := client.GETV1BmcStatusParams(selected(selection))

	res, err := apiclient.API.GETV1BmcStatus(command.Context(), params)
	if err != nil {
		return nil, apiclient.NewApiError(err)
	}
	return res, nil
}

func init() {
	statusTable = shared.RegisterTableFlags(statusCmd, statusColumns, "Host Name", "Manufacturer", "Locate LED", "Asset Tag", "Location", "CPU", "Memory", "Boot Order", "Boot Next", "OEM")

	bmcCmd.AddCommand(statusCmd)
}
