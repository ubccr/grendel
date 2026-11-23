// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	locationDryRun  bool
	locationTable   *shared.TableFlags
	locationFields  = map[string]*string{}
	locationColumns = []shared.Column{
		{Name: "Node"},
		{Name: "Current"},
		{Name: "New"},
		{Name: "State"},
		{Name: "Message"},
	}

	locationListTable   *shared.TableFlags
	locationListColumns = []shared.Column{
		{Name: "Node"},
		{Name: "Datacenter"},
		{Name: "Room"},
		{Name: "Row"},
		{Name: "Rack"},
		{Name: "Rack Unit"},
		{Name: "State"},
		{Name: "Message"},
	}

	locationCmd = &cobra.Command{
		Use:   "location",
		Short: "BMC rack location commands",
	}

	locationListCmd = &cobra.Command{
		Use:     "list <nodeset | all>",
		Aliases: []string{"ls"},
		Short:   "List the rack location each BMC reports",
		Args:    cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			res, err := bmcStatus(command, args[0])
			if err != nil {
				return err
			}

			t := locationListTable.Table()
			for _, result := range res.Results {
				loc := result.Data.Value.Location.Value
				ru := ""
				if unit, ok := loc.RackUnit.Get(); ok {
					ru = strconv.Itoa(unit)
				}
				t.AppendRow(
					result.Node,
					loc.Datacenter,
					loc.Room,
					loc.Row,
					loc.Rack,
					ru,
					colorState(result.Success),
					result.Message,
				)
			}
			t.Print()

			return summaryLine(summary(res.Summary))
		},
	}

	locationSetCmd = &cobra.Command{
		Use:   "set <nodeset | all>",
		Short: "Set rack locations from templates",
		Long: `Set the rack location each BMC records, from a Go template per field rendered for each node. A field without a flag uses its bmc.location default from the server config. Pass an empty value to leave a field as it is, or a template that renders empty, such as '{{ "" }}', to clear it.

Templates see .Name, .Parts (the name split on -), .Tags (key=value tags by key), .TagList, and .Index and .Count (the node's 0-based position among the selected nodes, in natural name order, and how many there are), with the functions split, join, trimPrefix, trimSuffix, trimLeft, replace, upper, lower, default, atoi, add, printf and tag "key". --rack-unit must render to a whole number, so "04" is fine.

On a Dell iDRAC this writes the ServerTopology attributes, and elsewhere the chassis Location.`,
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			req := &client.BmcLocationBody{DryRun: client.NewOptBool(locationDryRun)}
			flags := command.Flags()
			field := func(name string) client.OptNilString {
				if !flags.Changed(name) {
					return client.OptNilString{}
				}
				return client.NewOptNilString(*locationFields[name])
			}
			req.Datacenter = field("datacenter")
			req.Room = field("room")
			req.Row = field("row")
			req.Rack = field("rack")
			req.RackUnit = field("rack-unit")

			params := client.POSTV1BmcLocationParams(selected(args[0]))

			res, err := apiclient.API.POSTV1BmcLocation(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := locationTable.Table()
			for _, result := range res.Results {
				d := result.Data.Value
				t.AppendRow(
					result.Node,
					formatLocation(location(d.Current)),
					changedCell(formatLocation(location(d.Desired)), d.Changed),
					colorState(result.Success),
					result.Message,
				)
			}
			t.Print()

			if locationDryRun {
				fmt.Println("\ndry run, nothing was written")
			}

			return summaryLine(summary(res.Summary))
		},
	}
)

func init() {
	locationTable = shared.RegisterTableFlags(locationSetCmd, locationColumns)
	flags := locationSetCmd.Flags()
	flags.BoolVar(&locationDryRun, "dry-run", false, "read the BMCs and show what would change, without writing")
	for _, f := range []struct{ name, usage string }{
		{"datacenter", "template for the data center or building"},
		{"room", "template for the room"},
		{"row", "template for the row, which iDRAC calls the aisle"},
		{"rack", "template for the rack name"},
		{"rack-unit", "template for the rack unit, rendering to a whole number"},
	} {
		locationFields[f.name] = flags.String(f.name, "", f.usage)
	}

	locationListTable = shared.RegisterTableFlags(locationListCmd, locationListColumns)

	locationCmd.AddCommand(locationListCmd, locationSetCmd)
	bmcCmd.AddCommand(locationCmd)
}
