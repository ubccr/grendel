// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"github.com/ubccr/grendel/internal/settings"
	"github.com/ubccr/grendel/internal/tftp"
)

const (
	defaultTFTPListen = 6969
)

func init() {

	register(&Service{
		Name: "tftp",
		New:  newTFTP,
	})
}

func newTFTP() (Runner, error) {
	return tftp.NewServer(DB, getListenAddress("tftp.listen"), settings.TFTPFirmwareDir.Get())
}
