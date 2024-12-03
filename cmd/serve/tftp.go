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
	"github.com/ubccr/grendel/internal/tftp"
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
	tftpListen, err := GetListenAddress(viper.GetString("tftp.listen"))
	if err != nil {
		return err
	}

	tftpServer, err := tftp.NewServer(DB, tftpListen)
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
