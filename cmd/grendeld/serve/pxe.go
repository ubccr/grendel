// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"github.com/ubccr/grendel/internal/dhcp"
)

func init() {

	register(&Service{
		Name: "pxe",
		New:  newPXE,
	})
}

func newPXE() (Runner, error) {
	pxeListen := getListenAddress("pxe.listen")

	srv, err := dhcp.NewPXEServer(DB, pxeListen)
	if err != nil {
		return nil, err
	}

	return srv, nil
}
