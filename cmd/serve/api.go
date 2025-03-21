// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package serve

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/api"
	"gopkg.in/tomb.v2"
)

func init() {
	apiCmd.PersistentFlags().String("api-listen", fmt.Sprintf("127.0.0.1:%d", api.DefaultPort), "address to listen on")
	viper.BindPFlag("api.listen", apiCmd.PersistentFlags().Lookup("api-listen"))
	apiCmd.PersistentFlags().String("api-socket", "", "path to unix socket")
	viper.BindPFlag("api.socket_path", apiCmd.PersistentFlags().Lookup("api-socket"))
	apiCmd.PersistentFlags().String("api-cert", "", "path to ssl cert")
	viper.BindPFlag("api.cert", apiCmd.PersistentFlags().Lookup("api-cert"))
	apiCmd.PersistentFlags().String("api-key", "", "path to ssl key")
	viper.BindPFlag("api.key", apiCmd.PersistentFlags().Lookup("api-key"))

	serveCmd.AddCommand(apiCmd)
}

var (
	apiCmd = &cobra.Command{
		Use:   "api",
		Short: "Run API server",
		Long:  `Run API server`,
		RunE: func(command *cobra.Command, args []string) error {
			t := NewInterruptTomb()
			t.Go(func() error { return serveAPI(t) })
			return t.Wait()
		},
	}
)

func serveAPI(t *tomb.Tomb) error {
	apiListen, err := GetListenAddress(viper.GetString("api.listen"))
	if err != nil {
		return err
	}

	apiServer, err := api.NewServer(DB, viper.GetString("api.socket_path"), apiListen)
	if err != nil {
		return err
	}

	apiServer.KeyFile = viper.GetString("api.key")
	apiServer.CertFile = viper.GetString("api.cert")
	apiServer.CORS = viper.GetBool("api.cors")
	apiServer.SwaggerUI = viper.GetBool("api.swagger_ui")

	if viper.IsSet("api.listen") && !viper.IsSet("client.api_key") {
		cmd.Log.Warn("client.api_key is not set, CLI authentication will not work. Either bind the API to a unix socket or signup for an account in the web ui and create a token")
	}

	t.Go(func() error {
		time.Sleep(1 * time.Second)
		<-t.Dying()
		cmd.Log.Info("Shutting down API server...")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := apiServer.Shutdown(ctxShutdown); err != nil {
			cmd.Log.Errorf("Failed shutting down API server: %s", err)
			return err
		}

		return nil
	})

	return apiServer.Serve()
}
