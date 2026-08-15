// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	listFilter  []string
	listOptions *cmd.TableOptions
	listColumns = []cmd.Column{
		{Name: "Name"},
		{Name: "Kernel"},
		{Name: "Command Line", MinWidth: 10},
		{Name: "Initrd"},
		{Name: "Provision Templates"},
		{Name: "Verify"},
	}

	listCmd = &cobra.Command{
		Use:     "list [names]...",
		Short:   "List images",
		Aliases: []string{"ls"},
		Args:    cobra.ArbitraryArgs,
		RunE: func(command *cobra.Command, args []string) error {
			params := client.GETV1ImagesFindParams{
				Names: client.NewOptString(strings.Join(args, ",")),
			}
			images, err := cmd.API.GETV1ImagesFind(command.Context(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			t := cmd.NewTable(listColumns, listFilter).Options(*listOptions).Empty("-")

			for _, image := range images {
				var sb strings.Builder
				for k, v := range image.ProvisionTemplates.Value {
					if sb.Len() > 0 {
						sb.WriteString(",")
					}

					fmt.Fprintf(&sb, "%s=%s", k, v.Value)
				}

				t.AppendRow(image.Name, image.Kernel, image.Cmdline.Value, strings.Join(image.Initrd, ","), sb.String(), strconv.FormatBool(image.Verify.Value))
			}

			t.Print()

			return nil
		},
	}
)

func init() {
	listOptions = cmd.RegisterTableFlags(listCmd)
	listCmd.Flags().StringSliceVar(&listFilter, "filter", []string{"Command Line"}, "filter table columns by name")

	listCmd.RegisterFlagCompletionFunc("filter", cmd.ColumnCompletion(listColumns))

	imageCmd.AddCommand(listCmd)
}
