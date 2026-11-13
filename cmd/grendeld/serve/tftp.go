// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/tftp"
)

const (
	defaultTFTPListen = 6969
)

func init() {
	serveCmd.PersistentFlags().String("tftp-listen", "0.0.0.0:69", "address to listen on")
	viper.BindPFlag("tftp.listen", serveCmd.PersistentFlags().Lookup("tftp-listen"))
	serveCmd.PersistentFlags().String("tftp-firmware-dir", "", "directory of iPXE EFI builds to serve in place of the embedded ones, such as signed copies written by 'grendeld firmware export'")
	viper.BindPFlag("tftp.firmware_dir", serveCmd.PersistentFlags().Lookup("tftp-firmware-dir"))

	register(&Service{
		Name: "tftp",
		New:  newTFTP,
	})
}

func newTFTP() (Runner, error) {
	return tftp.NewServer(DB, getListenAddress("tftp.listen"), viper.GetString("tftp.firmware_dir"))
}
