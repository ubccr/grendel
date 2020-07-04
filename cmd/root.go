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

package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	golog "log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/api"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/util"
)

var (
	cfgFile     string
	cfgFileUsed string
	apiEndPoint string
	debug       bool
	verbose     bool

	Log  = logger.GetLogger("CLI")
	Root = &cobra.Command{
		Use:     "grendel",
		Version: api.Version,
		Short:   "Bare Metal Provisioning for HPC",
		Long:    ``,
	}
)

func Execute() {
	if err := Root.Execute(); err != nil {
		Log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	Root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	Root.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug messages")
	Root.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose messages")
	Root.PersistentFlags().String("endpoint", "grendel-api.socket", "Grendel API endpoint")
	viper.BindPFlag("client.api_endpoint", Root.PersistentFlags().Lookup("endpoint"))

	Root.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		return SetupLogging()
	}
}

func NewClient() (*client.APIClient, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: viper.GetBool("api.insecure")}}

	cacert := viper.GetString("client.cacert")
	pem, err := ioutil.ReadFile(cacert)
	if err == nil {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("Failed to read cacert: %s", cacert)
		}

		tr = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: certPool, InsecureSkipVerify: false}}
	}

	endpoint := viper.GetString("client.api_endpoint")

	// Is endpoint a path to a unix domain socket?
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		tr = &http.Transport{
			DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
				dialer := net.Dialer{}
				return dialer.DialContext(ctx, "unix", endpoint)
			},
		}
	}

	rclient := retryablehttp.NewClient()
	rclient.HTTPClient = &http.Client{Timeout: time.Second * 3600, Transport: tr}
	rclient.Logger = Log

	cfg := client.NewConfiguration()
	cfg.HTTPClient = rclient.StandardClient()

	client := client.NewAPIClient(cfg)

	return client, nil
}

func NewApiError(msg string, err error) error {
	var ge client.GenericOpenAPIError
	if errors.As(err, &ge) {
		apiErr := ge.Model()
		if apiErr != nil {
			return fmt.Errorf("%s: %s - %w", msg, apiErr.(client.ErrorResponse).Message, ge)
		}
	}

	return err
}

func SetupLogging() error {
	if debug {
		Log.Logger.SetLevel(logrus.DebugLevel)
	} else if verbose {
		Log.Logger.SetLevel(logrus.InfoLevel)
	} else {
		Log.Logger.SetLevel(logrus.WarnLevel)
	}
	golog.SetOutput(ioutil.Discard)

	if cfgFileUsed != "" {
		Log.Infof("Using config file: %s", cfgFileUsed)
	}

	Root.SilenceUsage = true
	Root.SilenceErrors = true

	return nil
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			Log.Fatal(err)
		}

		viper.AddConfigPath("/etc/grendel/")
		viper.AddConfigPath(home)
		viper.SetConfigName("grendel")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("grendel")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err == nil {
		cfgFileUsed = viper.ConfigFileUsed()
	}

	if !viper.IsSet("api.secret") {
		secret, err := util.GenerateSecret(32)
		if err != nil {
			Log.Fatal(err)
		}

		viper.Set("api.secret", secret)
	}
}
