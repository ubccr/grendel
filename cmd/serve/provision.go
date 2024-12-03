// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/provision"
	"gopkg.in/tomb.v2"
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
			t := NewInterruptTomb()
			t.Go(func() error { return serveProvision(t) })
			return t.Wait()
		},
	}
)

func serveProvision(t *tomb.Tomb) error {
	pListen, err := GetListenAddress(viper.GetString("provision.listen"))
	if err != nil {
		return err
	}

	srv, err := provision.NewServer(DB, pListen)
	if err != nil {
		return err
	}

	srv.KeyFile = viper.GetString("provision.key")
	srv.CertFile = viper.GetString("provision.cert")
	srv.RepoDir = viper.GetString("provision.repo_dir")

	t.Go(func() error {
		time.Sleep(1 * time.Second)
		<-t.Dying()
		cmd.Log.Info("Shutting down Provision server...")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctxShutdown); err != nil {
			cmd.Log.Errorf("Failed shutting down Provision server: %s", err)
			return err
		}

		return nil
	})

	return srv.Serve(viper.GetString("provision.default_image"))
}
