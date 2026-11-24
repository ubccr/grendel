// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"errors"

	"github.com/ubccr/grendel/internal/api"
	"github.com/ubccr/grendel/internal/settings"
)

func init() {

	register(&Service{
		Name: "api",
		New:  newAPI,
	})
}

func newAPI() (Runner, error) {
	socket := settings.APISocketPath.Get()
	tcpListen := getListenAddress("api.listen")

	if socket == "" && tcpListen == "" {
		return nil, errors.New("set api.socket_path or api.listen")
	}

	srv, err := api.NewServer(DB, socket, tcpListen)
	if err != nil {
		return nil, err
	}

	srv.KeyFile = settings.APIKey.Get()
	srv.CertFile = settings.APICert.Get()
	srv.CORS = settings.APICORS.Get()
	srv.SwaggerUI = settings.APISwaggerUI.Get()

	return srv, nil
}
