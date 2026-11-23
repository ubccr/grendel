// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	biosDescribeSearch  string
	biosDescribeTable   *shared.TableFlags
	biosDescribeColumns = []shared.Column{
		{Name: "Attribute"},
		{Name: "Name"},
		{Name: "Menu"},
		{Name: "Type"},
		{Name: "Allowed"},
		{Name: "Default"},
		{Name: "Flags"},
		{Name: "Help"},
	}
	biosDescribeCmd = &cobra.Command{
		Use:   "describe <node> [attribute]...",
		Short: "Describe BIOS attributes from the BMC's attribute registry",
		Long:  "Describe BIOS attributes from the BMC's attribute registry: their display names, types, allowed values and whether they can be changed",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			params := client.GETV1BmcBiosRegistryParams{
				Nodeset: client.NewOptString(args[0]),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			// the route answers for exactly one node, and refuses a nodeset that names more
			res, err := apiclient.API.GETV1BmcBiosRegistry(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}
			if len(res.Results) == 0 {
				return errors.New("no result returned")
			}

			result := res.Results[0]
			if !result.Success {
				return fmt.Errorf("%s: %s", result.Node, result.Message)
			}
			reg := result.Data.Value
			if !reg.Available {
				return fmt.Errorf("%s publishes no attribute registry it serves itself, 'bmc bios get' still lists the attributes by name", result.Node)
			}

			search := strings.ToLower(biosDescribeSearch)
			t := biosDescribeTable.Table()
			for _, attr := range reg.Attributes {
				name := attr.Name
				if len(args) > 1 && !slices.Contains(args[1:], name) {
					continue
				}
				if search != "" && !strings.Contains(strings.ToLower(name+" "+attr.DisplayName+" "+attr.HelpText), search) {
					continue
				}

				t.AppendRow(
					name,
					attr.DisplayName,
					strings.TrimPrefix(attr.MenuPath, "./"),
					attr.Type,
					allowedValues(attr),
					rawValue(attr.DefaultValue),
					attributeFlags(attr),
					attr.HelpText,
				)
			}
			t.Print()
			return nil
		},
	}
)

func allowedValues(attr client.BmcResponseModelBmcBiosRegistryResultsItemDataAttributesItem) string {
	if len(attr.Values) > 0 {
		names := make([]string, 0, len(attr.Values))
		for _, v := range attr.Values {
			names = append(names, v.Name)
		}
		return strings.Join(names, "\n")
	}

	lower, hasLower := attr.LowerBound.Get()
	upper, hasUpper := attr.UpperBound.Get()
	if hasLower || hasUpper {
		return fmt.Sprintf("%s..%s", optBound(lower, hasLower), optBound(upper, hasUpper))
	}

	if maxLen, ok := attr.MaxLength.Get(); ok {
		minLen, hasMin := attr.MinLength.Get()
		return fmt.Sprintf("%s to %d characters", optBound(int64(minLen), hasMin), maxLen)
	}
	return ""
}

func optBound(v int64, set bool) string {
	if !set {
		return ""
	}
	return strconv.FormatInt(v, 10)
}

func attributeFlags(attr client.BmcResponseModelBmcBiosRegistryResultsItemDataAttributesItem) string {
	var flags []string
	for _, f := range []struct {
		on   bool
		name string
	}{
		{attr.ReadOnly, "read-only"},
		{attr.Immutable, "immutable"},
		{attr.GrayOut, "grayed out"},
		{attr.Hidden, "hidden"},
		{attr.WriteOnly, "write-only"},
		{attr.ResetRequired, "needs reboot"},
		{attr.SystemUnique, "per system"},
	} {
		if f.on {
			flags = append(flags, f.name)
		}
	}
	return strings.Join(flags, ", ")
}

func init() {
	biosDescribeTable = shared.RegisterTableFlags(biosDescribeCmd, biosDescribeColumns, "Help")
	biosDescribeCmd.Flags().StringVar(&biosDescribeSearch, "search", "", "show only attributes whose name, display name or help mentions this text")

	biosCmd.AddCommand(biosDescribeCmd)
}
