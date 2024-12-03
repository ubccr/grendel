// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package host

import (
	"context"
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/model"
)

var (
	importCmd = &cobra.Command{
		Use:   "import",
		Short: "import hosts",
		Long:  `import hosts`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			for _, name := range args {
				file, err := os.Open(name)
				if err != nil {
				}
				defer file.Close()

				cmd.Log.Infof("Processing file: %s", name)

				var hosts model.HostList
				if err := json.NewDecoder(file).Decode(&hosts); err != nil {
					return err
				}

				_, err = gc.HostApi.StoreHosts(context.Background(), hosts)
				if err != nil {
					return cmd.NewApiError("Failed to store hosts", err)
				}

				cmd.Log.Infof("Successfully imported %d hosts from: %s", len(hosts), name)
			}
			return nil

		},
	}
)

func init() {
	hostCmd.AddCommand(importCmd)
}
