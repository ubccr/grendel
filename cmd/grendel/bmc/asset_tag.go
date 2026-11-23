// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	assetTagDryRun       bool
	assetTagSerialColumn string
	assetTagTagColumn    string
	assetTagSetTable     *shared.TableFlags
	assetTagImportTable  *shared.TableFlags
	assetTagColumns      = []shared.Column{
		{Name: "Node"},
		{Name: "Serial"},
		{Name: "Current"},
		{Name: "New"},
		{Name: "State"},
		{Name: "Message"},
	}

	assetTagListTable   *shared.TableFlags
	assetTagListColumns = []shared.Column{
		{Name: "Node"},
		{Name: "Serial"},
		{Name: "Asset Tag"},
		{Name: "State"},
		{Name: "Message"},
	}

	assetTagCmd = &cobra.Command{
		Use:   "asset-tag",
		Short: "BMC asset tag commands",
	}

	assetTagListCmd = &cobra.Command{
		Use:     "list <nodeset | all>",
		Aliases: []string{"ls"},
		Short:   "List the asset tag and serial each BMC reports",
		Args:    cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			res, err := bmcStatus(command, args[0])
			if err != nil {
				return err
			}

			t := assetTagListTable.Table()
			for _, result := range res.Results {
				o := result.Data.Value
				t.AppendRow(
					result.Node,
					o.SerialNumber,
					o.AssetTag,
					colorState(result.Success),
					result.Message,
				)
			}
			t.Print()

			return summaryLine(summary(res.Summary))
		},
	}

	assetTagSetCmd = &cobra.Command{
		Use:   "set <nodeset | all> [template]",
		Short: "Set asset tags from a template",
		Long: `Set the asset tag of each node from a Go template, rendered per node. A plain string is a valid template. Without a template, bmc.asset_tag from the server config is used.

Templates see .Name, .Parts (the name split on -), .Tags (key=value tags by key), .TagList, and .Index and .Count (the node's 0-based position among the selected nodes, in natural name order, and how many there are), with the functions split, join, trimPrefix, trimSuffix, trimLeft, replace, upper, lower, default, atoi, add, printf and tag "key".`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(command *cobra.Command, args []string) error {
			req := &client.BmcAssetTagBody{DryRun: client.NewOptBool(assetTagDryRun)}
			if len(args) == 2 {
				req.Template = client.NewOptNilString(args[1])
			}
			return runAssetTag(command, args[0], req, nil, assetTagSetTable)
		},
	}

	assetTagImportCmd = &cobra.Command{
		Use:   "import <file.csv> <nodeset | all>",
		Short: "Set asset tags from a CSV of serial numbers",
		Long: `Set asset tags from a CSV that pairs serial numbers (Dell service tags) with asset tags. Each node is matched by the serial its BMC reports, and a selected node whose serial is not in the file fails.

The columns are found by a header row (serial, serial number or service tag, and asset tag or tag), or picked with --serial-column and --tag-column. A file with no header row is read as serial then asset tag.`,
		Args: cobra.ExactArgs(2),
		RunE: func(command *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			rows, err := parseAssetCSV(f, assetTagSerialColumn, assetTagTagColumn)
			if err != nil {
				return err
			}

			bySerial := client.BmcAssetTagBodyBySerial{}
			for serial, tag := range rows {
				bySerial[serial] = client.NewNilString(tag)
			}
			req := &client.BmcAssetTagBody{
				BySerial: client.NewOptNilBmcAssetTagBodyBySerial(bySerial),
				DryRun:   client.NewOptBool(assetTagDryRun),
			}
			return runAssetTag(command, args[1], req, rows, assetTagImportTable)
		},
	}
)

// runAssetTag sends the request and prints the table. rows is the CSV, when there is one, so the rows no node took can be listed
func runAssetTag(command *cobra.Command, selection string, req *client.BmcAssetTagBody, rows map[string]string, table *shared.TableFlags) error {
	params := client.POSTV1BmcAssetTagParams(selected(selection))

	res, err := apiclient.API.POSTV1BmcAssetTag(command.Context(), req, params)
	if err != nil {
		return apiclient.NewApiError(err)
	}

	applied := map[string]bool{}
	t := table.Table()
	for _, result := range res.Results {
		d := result.Data.Value
		applied[d.Serial] = true
		t.AppendRow(
			result.Node,
			d.Serial,
			d.Current,
			changedCell(d.Desired, d.Changed),
			colorState(result.Success),
			result.Message,
		)
	}
	t.Print()

	if rows != nil {
		var unused []string
		for serial := range rows {
			if !applied[serial] {
				unused = append(unused, serial)
			}
		}
		if len(unused) > 0 {
			slices.Sort(unused)
			fmt.Printf("\n%d of %d CSV row(s) went to no node: %s\n", len(unused), len(rows), strings.Join(unused, ", "))
		}
	}
	if assetTagDryRun {
		fmt.Println("\ndry run, nothing was written")
	}

	return summaryLine(summary(res.Summary))
}

func init() {
	assetTagSetTable = shared.RegisterTableFlags(assetTagSetCmd, assetTagColumns)
	assetTagImportTable = shared.RegisterTableFlags(assetTagImportCmd, assetTagColumns)
	for _, c := range []*cobra.Command{assetTagSetCmd, assetTagImportCmd} {
		c.Flags().BoolVar(&assetTagDryRun, "dry-run", false, "read the BMCs and show what would change, without writing")
	}

	assetTagImportCmd.Flags().StringVar(&assetTagSerialColumn, "serial-column", "", "header of the serial number column")
	assetTagImportCmd.Flags().StringVar(&assetTagTagColumn, "tag-column", "", "header of the asset tag column")

	assetTagListTable = shared.RegisterTableFlags(assetTagListCmd, assetTagListColumns)

	assetTagCmd.AddCommand(assetTagListCmd, assetTagSetCmd, assetTagImportCmd)
	bmcCmd.AddCommand(assetTagCmd)
}
