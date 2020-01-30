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
	"github.com/ubccr/grendel/provision"
)

func init() {
	provisionCmd.Flags().String("provision-listen", "0.0.0.0:80", "address to listen on")
	viper.BindPFlag("provision.listen", provisionCmd.Flags().Lookup("provision-listen"))
	provisionCmd.Flags().String("provision-cert", "", "path to ssl cert")
	viper.BindPFlag("provision.cert", provisionCmd.Flags().Lookup("provision-cert"))
	provisionCmd.Flags().String("provision-key", "", "path to ssl key")
	viper.BindPFlag("provision.key", provisionCmd.Flags().Lookup("provision-key"))
	provisionCmd.Flags().String("default-image", "", "default image name")
	viper.BindPFlag("provision.default_image", provisionCmd.Flags().Lookup("default-image"))
	provisionCmd.Flags().String("repo-dir", "", "path to repo dir")
	viper.BindPFlag("provision.repo_dir", provisionCmd.Flags().Lookup("repo-dir"))

	serveCmd.AddCommand(provisionCmd)
}

var (
	provisionCmd = &cobra.Command{
		Use:   "provision",
		Short: "Run Provision server",
		Long:  `Run Provision server`,
		RunE: func(command *cobra.Command, args []string) error {
			ctx, _ := NewInterruptContext()
			return serveProvision(ctx)
		},
	}
)

func serveProvision(ctx context.Context) error {
	srv, err := provision.NewServer(DB, viper.GetString("provision.listen"))
	if err != nil {
		return err
	}

	srv.KeyFile = viper.GetString("provision.key")
	srv.CertFile = viper.GetString("provision.cert")
	srv.RepoDir = viper.GetString("provision.repo_dir")

	if err := srv.Serve(ctx, viper.GetString("provision.default_image")); err != nil {
		return err
	}

	return nil
}
