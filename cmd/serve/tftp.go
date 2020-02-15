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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/tftp"
	"gopkg.in/tomb.v2"
)

func init() {
	tftpCmd.PersistentFlags().String("tftp-listen", "0.0.0.0:69", "address to listen on")
	viper.BindPFlag("tftp.listen", tftpCmd.PersistentFlags().Lookup("tftp-listen"))

	serveCmd.AddCommand(tftpCmd)
}

var (
	tftpCmd = &cobra.Command{
		Use:   "tftp",
		Short: "Run TFTP server",
		Long:  `Run TFTP server`,
		RunE: func(command *cobra.Command, args []string) error {
			t := NewInterruptTomb()
			t.Go(func() error { return serveTFTP(t) })
			return t.Wait()
		},
	}
)

func serveTFTP(t *tomb.Tomb) error {
	tftpServer, err := tftp.NewServer(viper.GetString("tftp.listen"))
	if err != nil {
		return err
	}

	t.Go(tftpServer.Serve)

	t.Go(func() error {
		time.Sleep(1 * time.Second)
		<-t.Dying()
		cmd.Log.Info("Shutting down TFTP server...")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := tftpServer.Shutdown(ctxShutdown); err != nil {
			cmd.Log.Errorf("Failed shutting down TFTP server: %s", err)
			return err
		}

		return nil
	})

	return nil
}
