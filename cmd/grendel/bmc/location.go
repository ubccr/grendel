// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	locationFilter  []string
	locationDryRun  bool
	locationOptions *shared.TableOptions
	locationFields  = map[string]*string{}
	locationColumns = []shared.Column{
		{Name: "Node"},
		{Name: "Current"},
		{Name: "New"},
		{Name: "State"},
		{Name: "Message"},
	}

	locationListFilter  []string
	locationListOptions *shared.TableOptions
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
		Example: `  grendel bmc location list cpn-f01-[01-40]
  grendel bmc location list all --filter Datacenter,Room,Message`,
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			res, err := bmcStatus(command, args[0])
			if err != nil {
				return err
			}

			t := shared.NewTable(locationListColumns, locationListFilter).Options(*locationListOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				loc := result.Data.Value.Location.Value
				ru := ""
				if loc.RackUnit.Set && !loc.RackUnit.Null {
					ru = strconv.Itoa(loc.RackUnit.Value)
				}
				t.AppendRow(
					result.Node.Value,
					loc.Datacenter.Value,
					loc.Room.Value,
					loc.Row.Value,
					loc.Rack.Value,
					ru,
					colorState(result.Success.Value),
					result.Message.Value,
				)
			}
			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}

	locationSetCmd = &cobra.Command{
		Use:   "set <nodeset | all>",
		Short: "Set rack locations from templates",
		Long: `Set the rack location each BMC records, from a Go template per field rendered for each node. A field without a flag uses its bmc.location default from the server config. Pass an empty value to leave a field as it is, or a template that renders empty, such as '{{ "" }}', to clear it.

Templates see .Name, .Parts (the name split on -), .Tags (key=value tags by key), .TagList, and .Index and .Count (the node's 0-based position among the selected nodes, in natural name order, and how many there are), with the functions split, join, trimPrefix, trimSuffix, trimLeft, replace, upper, lower, default, atoi, add, printf and tag "key". --rack-unit must render to a whole number, so "04" is fine.

On a Dell iDRAC this writes the ServerTopology attributes, and elsewhere the chassis Location.`,
		Example: `  grendel bmc location set cpn-f01-[01-40] --rack '{{ index .Parts 1 }}' --rack-unit '{{ index .Parts 2 }}' --dry-run
  grendel bmc location set all --room 'Suite 3' --datacenter VTC
  grendel bmc location set all --tags gpu`,
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

			nodeset, all := selector(args[0])
			params := client.POSTV1BmcLocationParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcLocation(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(locationColumns, locationFilter).Options(*locationOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				d := result.Data.Value
				cur, want := d.Current.Value, d.Desired.Value
				t.AppendRow(
					result.Node.Value,
					formatLocation(cur.Datacenter.Value, cur.Room.Value, cur.Row.Value, cur.Rack.Value, cur.RackUnit),
					changedCell(formatLocation(want.Datacenter.Value, want.Room.Value, want.Row.Value, want.Rack.Value, want.RackUnit), d.Changed.Value),
					colorState(result.Success.Value),
					result.Message.Value,
				)
			}
			t.Print()

			if locationDryRun {
				fmt.Println("\ndry run, nothing was written")
			}

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

func init() {
	locationOptions = shared.RegisterTableFlags(locationSetCmd)
	flags := locationSetCmd.Flags()
	flags.StringSliceVar(&locationFilter, "filter", nil, "filter table columns by name")
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
	locationSetCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(locationColumns))

	locationListOptions = shared.RegisterTableFlags(locationListCmd)
	locationListCmd.Flags().StringSliceVar(&locationListFilter, "filter", nil, "filter table columns by name")
	locationListCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(locationListColumns))

	locationCmd.AddCommand(locationListCmd, locationSetCmd)
	bmcCmd.AddCommand(locationCmd)
}
