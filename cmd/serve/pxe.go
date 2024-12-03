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
	"github.com/ubccr/grendel/internal/dhcp"
	"gopkg.in/tomb.v2"
)

func init() {
	pxeCmd.PersistentFlags().String("pxe-listen", "0.0.0.0:4011", "address to listen on")
	viper.BindPFlag("pxe.listen", pxeCmd.PersistentFlags().Lookup("pxe-listen"))

	serveCmd.AddCommand(pxeCmd)
}

var (
	pxeCmd = &cobra.Command{
		Use:   "pxe",
		Short: "Run DHCP PXE Boot server",
		Long:  `Run DHCP PXE Boot server`,
		RunE: func(command *cobra.Command, args []string) error {
			t := NewInterruptTomb()
			t.Go(func() error { return servePXE(t) })
			return t.Wait()
		},
	}
)

func servePXE(t *tomb.Tomb) error {
	pxeListen, err := GetListenAddress(viper.GetString("pxe.listen"))
	if err != nil {
		return err
	}

	srv, err := dhcp.NewPXEServer(DB, pxeListen)
	if err != nil {
		return err
	}

	t.Go(func() error {
		time.Sleep(1 * time.Second)
		<-t.Dying()
		cmd.Log.Info("Shutting down PXE server...")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctxShutdown); err != nil {
			cmd.Log.Errorf("Failed shutting down PXE server: %s", err)
			return err
		}

		return nil
	})

	return srv.Serve()
}
