// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	listCmd = &cobra.Command{
		Use:   "list {names... | all}",
		Short: "List images",
		Long:  `List images`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			var res []client.BootImage
			var err error

			if strings.ToLower(args[0]) == "all" {
				params := client.GETV1ImagesParams{}
				res, err = cmd.API.GETV1Images(command.Context(), params)
				if err != nil {
					return cmd.NewApiError(err)
				}
			} else {
				params := client.GETV1ImagesFindParams{
					Names: client.NewOptString(strings.Join(args, ",")),
				}
				res, err = cmd.API.GETV1ImagesFind(command.Context(), params)
				if err != nil {
					return cmd.NewApiError(err)
				}

			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

			fmt.Fprintln(w, "Name\tKernel\tInitrd\tLive Image\tProvision Templates")
			for _, image := range res {
				provisionTemplates := make([]string, 0)
				for k, v := range image.ProvisionTemplates.Value {
					provisionTemplates = append(provisionTemplates, fmt.Sprintf("\"%s\": \"%s\"", k, v.Value))
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", image.Name, image.Kernel, strings.Join(image.Initrd, ","), strings.Join(provisionTemplates, ","))
			}

			return w.Flush()
		},
	}
)

func init() {
	imageCmd.AddCommand(listCmd)
}
