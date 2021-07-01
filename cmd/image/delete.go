// Copyright 2021 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package image

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
)

var (
	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete images",
		Long:  `Delete images`,
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			_, err = gc.ImageApi.ImageDelete(context.Background(), args[0])
			if err != nil {
				return cmd.NewApiError("Failed to delete hosts", err)
			}

			fmt.Println("Successfully deleted image")

			return nil

		},
	}
)

func init() {
	imageCmd.AddCommand(deleteCmd)
}
