// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"errors"

	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/api"
)

func init() {
	serveCmd.PersistentFlags().String("api-listen", "", "address to listen on")
	viper.BindPFlag("api.listen", serveCmd.PersistentFlags().Lookup("api-listen"))
	serveCmd.PersistentFlags().String("api-socket", "", "path to unix socket")
	viper.BindPFlag("api.socket_path", serveCmd.PersistentFlags().Lookup("api-socket"))
	serveCmd.PersistentFlags().String("api-cert", "", "path to ssl cert")
	viper.BindPFlag("api.cert", serveCmd.PersistentFlags().Lookup("api-cert"))
	serveCmd.PersistentFlags().String("api-key", "", "path to ssl key")
	viper.BindPFlag("api.key", serveCmd.PersistentFlags().Lookup("api-key"))

	register(&Service{
		Name: "api",
		New:  newAPI,
	})
}

func newAPI() (Runner, error) {
	socket := viper.GetString("api.socket_path")
	tcpListen := getListenAddress("api.listen")

	if socket == "" && tcpListen == "" {
		return nil, errors.New("set api.socket_path or api.listen")
	}

	srv, err := api.NewServer(DB, socket, tcpListen)
	if err != nil {
		return nil, err
	}

	srv.KeyFile = viper.GetString("api.key")
	srv.CertFile = viper.GetString("api.cert")
	srv.CORS = viper.GetBool("api.cors")
	srv.SwaggerUI = viper.GetBool("api.swagger_ui")

	return srv, nil
}
