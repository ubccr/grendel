// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"github.com/ubccr/grendel/internal/provision"
	"github.com/ubccr/grendel/internal/settings"
)

func init() {

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

	srv.KeyFile = settings.ProvisionKey.Get()
	srv.CertFile = settings.ProvisionCert.Get()
	srv.RepoDir = settings.ProvisionRepoDir.Get()
	srv.TemplatesDir = settings.ProvisionTemplatesDir.Get()

	return &provisionRunner{Server: srv, defaultImage: settings.ProvisionDefaultImage.Get()}, nil
}
