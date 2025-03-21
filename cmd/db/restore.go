// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package db

import (
	"context"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	confirm    bool
	restoreCmd = &cobra.Command{
		Use:   "restore <filename>...",
		Short: "Restore database",
		Long:  `Restore database`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			for _, name := range args {
				file, err := os.ReadFile(name)
				if err != nil {
					fmt.Printf("failed to open file. name=%s err=%s", name, err)
				}

				// old data check
				gr := gjson.GetBytes(file, "hosts")
				for _, node := range gr.Array() {
					if node.Get("id").Type == gjson.String {
						return fmt.Errorf(`Old Grendel JSON schema detected. Please change the following fields: "id" -> "uid", "_id" -> "id". Example correct schema: "id": 0, "uid": "2TfJcVLhjDZkIrLpyFsbv68yzKF"`)
					}
				}
				gri := gjson.GetBytes(file, "images")
				for _, image := range gri.Array() {
					if image.Get("id").Type == gjson.String {
						return fmt.Errorf(`Old Grendel JSON schema detected. Please change the following fields: "id" -> "uid", "_id" -> "id". Example correct schema: "id": 0, "uid": "2TfJcVLhjDZkIrLpyFsbv68yzKF"`)
					}
				}
				var dump client.DataDump
				err = dump.UnmarshalJSON(file)
				if err != nil {
					return fmt.Errorf("failed to decode json: %w", err)
				}

				prompt := promptui.Prompt{
					Label:     "WARNING: database will be restored. Are you sure?",
					IsConfirm: true,
				}

				if !confirm {
					_, err = prompt.Run()
					if err != nil {
						fmt.Println("Restore cancelled.")
						return nil
					}
				}

				params := client.POSTV1DbRestoreParams{}
				res, err := gc.POSTV1DbRestore(context.Background(), &dump, params)
				if err != nil {
					return cmd.NewApiError(err)
				}

				err = cmd.NewApiResponse(res)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			return nil
		},
	}
)

func init() {
	dbCmd.AddCommand(restoreCmd)
	restoreCmd.PersistentFlags().BoolVarP(&confirm, "yes-i-really-mean-it", "y", false, "override yes prompt")
}
