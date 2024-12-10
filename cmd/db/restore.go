// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package db

import (
	"context"
	"encoding/json"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/model"
)

var (
	restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore database",
		Long:  `Restore database`,
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

				var dump model.DataDump
				if err := json.NewDecoder(file).Decode(&dump); err != nil {
					return err
				}

				prompt := promptui.Prompt{
					Label:     "WARNING: database will be restored. Are you sure?",
					IsConfirm: true,
				}

				_, err = prompt.Run()

				if err != nil {
					cmd.Log.Warn("Restore cancelled.")
					return nil
				}

				_, err = gc.RestoreApi.Restore(context.Background(), dump)
				if err != nil {
					return cmd.NewApiError("Failed to restore database", err)
				}

				cmd.Log.Warnf("Successfully restored database: %d users, %d hosts, %d images", len(dump.Users), len(dump.Hosts), len(dump.Images))
			}
			return nil
		},
	}
)

func init() {
	dbCmd.AddCommand(restoreCmd)
}
