// Copyright 2019 Grendel Authors. All rights reserved.
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

package serve

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/dhcp"
)

func init() {
	pxeCmd.PersistentFlags().String("pxe-listen", "0.0.0.0:4011", "address to listen on")
	viper.BindPFlag("pxe.listen", pxeCmd.PersistentFlags().Lookup("pxe-listen"))

	serveCmd.AddCommand(pxeCmd)
}

var (
	pxeCmd = &cobra.Command{
		Use:   "pxe",
		Short: "Run DHCP PXE Boot server",
		Long:  `Run DHCP PXE Boot server`,
		RunE: func(command *cobra.Command, args []string) error {
			ctx, _ := NewInterruptContext()
			return servePXE(ctx)
		},
	}
)

func servePXE(ctx context.Context) error {
	srv, err := dhcp.NewPXEServer(DB, viper.GetString("pxe.listen"))
	if err != nil {
		return err
	}

	if err := srv.Serve(ctx); err != nil {
		return err
	}

	return nil
}
