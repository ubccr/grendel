// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/provision"
)

func init() {
	serveCmd.PersistentFlags().String("provision-listen", "0.0.0.0:80", "address to listen on")
	viper.BindPFlag("provision.listen", serveCmd.PersistentFlags().Lookup("provision-listen"))
	serveCmd.PersistentFlags().String("provision-cert", "", "path to ssl cert")
	viper.BindPFlag("provision.cert", serveCmd.PersistentFlags().Lookup("provision-cert"))
	serveCmd.PersistentFlags().String("provision-key", "", "path to ssl key")
	viper.BindPFlag("provision.key", serveCmd.PersistentFlags().Lookup("provision-key"))
	serveCmd.PersistentFlags().String("default-image", "", "default image name")
	viper.BindPFlag("provision.default_image", serveCmd.PersistentFlags().Lookup("default-image"))
	serveCmd.PersistentFlags().String("repo-dir", "", "path to repo dir")
	viper.BindPFlag("provision.repo_dir", serveCmd.PersistentFlags().Lookup("repo-dir"))
	serveCmd.PersistentFlags().String("templates-dir", "", "path to templates dir")
	viper.BindPFlag("provision.templates_dir", serveCmd.PersistentFlags().Lookup("templates-dir"))

	register(&Service{
		Name: "provision",
		New:  newProvision,
	})
}

// provisionRunner adapts the provision server, whose Serve takes the default image name, to the Runner interface
type provisionRunner struct {
	*provision.Server
	defaultImage string
}

func (p *provisionRunner) Serve() error {
	return p.Server.Serve(p.defaultImage)
}

func newProvision() (Runner, error) {
	pListen := getListenAddress("provision.listen")

	srv, err := provision.NewServer(DB, pListen)
	if err != nil {
		return nil, err
	}

	srv.KeyFile = viper.GetString("provision.key")
	srv.CertFile = viper.GetString("provision.cert")
	srv.RepoDir = viper.GetString("provision.repo_dir")
	srv.TemplatesDir = viper.GetString("provision.templates_dir")

	return &provisionRunner{Server: srv, defaultImage: viper.GetString("provision.default_image")}, nil
}
