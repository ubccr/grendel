// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/dhcp"
)

func init() {
	serveCmd.PersistentFlags().String("pxe-listen", "0.0.0.0:4011", "address to listen on")
	viper.BindPFlag("pxe.listen", serveCmd.PersistentFlags().Lookup("pxe-listen"))

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
