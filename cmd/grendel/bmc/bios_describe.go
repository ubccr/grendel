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
	biosDescribeFilter  []string
	biosDescribeSearch  string
	biosDescribeOptions *shared.TableOptions
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

			res, err := apiclient.API.GETV1BmcBiosRegistry(command.Context(), params)
			if err != nil {
				return apiclient.NewApiError(err)
			}
			if len(res.Results.Value) == 0 {
				return errors.New("no result returned")
			}

			result := res.Results.Value[0].Value
			if !result.Success.Value {
				return fmt.Errorf("%s: %s", result.Node.Value, result.Message.Value)
			}
			reg := result.Data.Value
			if !reg.Available.Value {
				return fmt.Errorf("%s publishes no attribute registry it serves itself, 'bmc bios get' still lists the attributes by name", result.Node.Value)
			}

			search := strings.ToLower(biosDescribeSearch)
			t := shared.NewTable(biosDescribeColumns, biosDescribeFilter).Options(*biosDescribeOptions).Empty("-")
			for _, a := range reg.Attributes {
				attr := a
				name := attr.Name.Value
				if len(args) > 1 && !slices.Contains(args[1:], name) {
					continue
				}
				if search != "" && !strings.Contains(strings.ToLower(name+" "+attr.DisplayName.Value+" "+attr.HelpText.Value), search) {
					continue
				}

				t.AppendRow(
					name,
					attr.DisplayName.Value,
					strings.TrimPrefix(attr.MenuPath.Value, "./"),
					attr.Type.Value,
					allowedValues(attr),
					rawValue(attr.DefaultValue),
					attributeFlags(attr),
					attr.HelpText.Value,
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
			names = append(names, v.Name.Value)
		}
		return strings.Join(names, "\n")
	}

	lower, upper := attr.LowerBound, attr.UpperBound
	if lower.Set && !lower.Null || upper.Set && !upper.Null {
		return fmt.Sprintf("%s..%s", optBound(lower.Value, lower.Set && !lower.Null), optBound(upper.Value, upper.Set && !upper.Null))
	}

	minLen, maxLen := attr.MinLength, attr.MaxLength
	if maxLen.Set && !maxLen.Null {
		return fmt.Sprintf("%s to %d characters", optBound(int64(minLen.Value), minLen.Set && !minLen.Null), maxLen.Value)
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
		{attr.ReadOnly.Value, "read-only"},
		{attr.Immutable.Value, "immutable"},
		{attr.GrayOut.Value, "grayed out"},
		{attr.Hidden.Value, "hidden"},
		{attr.WriteOnly.Value, "write-only"},
		{attr.ResetRequired.Value, "needs reboot"},
		{attr.SystemUnique.Value, "per system"},
	} {
		if f.on {
			flags = append(flags, f.name)
		}
	}
	return strings.Join(flags, ", ")
}

func init() {
	biosDescribeOptions = shared.RegisterTableFlags(biosDescribeCmd)
	biosDescribeCmd.Flags().StringSliceVar(&biosDescribeFilter, "filter", []string{"Help"}, "filter table columns by name")
	biosDescribeCmd.Flags().StringVar(&biosDescribeSearch, "search", "", "show only attributes whose name, display name or help mentions this text")
	biosDescribeCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(biosDescribeColumns))

	biosCmd.AddCommand(biosDescribeCmd)
}
