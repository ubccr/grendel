// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

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
	"os"
	"strings"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/api"
	"github.com/ubccr/grendel/internal/config"
	"github.com/ubccr/grendel/internal/logger"
	"github.com/ubccr/grendel/internal/util"
	"github.com/ubccr/grendel/pkg/client"
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

type ogenAuth struct{}

func newAuthHandler() ogenAuth {
	return ogenAuth{}
}

func (o ogenAuth) HeaderAuth(ctx context.Context, operationName string, c *client.Client) (client.HeaderAuth, error) {
	auth := client.HeaderAuth{Token: viper.GetString("client.api_key")}
	return auth, nil
}

func (o ogenAuth) CookieAuth(ctx context.Context, operationName string, c *client.Client) (client.CookieAuth, error) {
	auth := client.CookieAuth{Token: viper.GetString("client.api_key")}
	return auth, nil
}

func NewOgenClient() (*client.Client, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: viper.GetBool("client.insecure")}}

	cacert := viper.GetString("client.cacert")
	pem, err := os.ReadFile(cacert)
	if err == nil {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("Failed to read cacert: %s", cacert)
		}

		tr = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: certPool, InsecureSkipVerify: false}}
	}
	endpoint := viper.GetString("client.api_endpoint")
	if !strings.HasPrefix(endpoint, "http") {
		tr = &http.Transport{
			DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
				dialer := net.Dialer{}
				return dialer.DialContext(ctx, "unix", viper.GetString("client.api_endpoint"))
			},
		}
		endpoint = "http://localhost"
	}

	rclient := retryablehttp.NewClient()
	rclient.HTTPClient = &http.Client{Timeout: time.Second * 3600, Transport: tr}
	rclient.Logger = Log
	httpClient := rclient.StandardClient()
	client, err := client.NewClient(endpoint, newAuthHandler(), client.WithClient(httpClient))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewApiError(apiError error) error {
	var t *client.HTTPErrorStatusCode

	if !errors.As(apiError, &t) {
		return apiError
	}

	httpError := t.GetResponse()

	return fmt.Errorf("%d: %s - %s", httpError.GetStatus().Value, httpError.GetTitle().Value, httpError.GetDetail().Value)
}
func NewApiResponse(res *client.GenericResponse) error {
	fmt.Printf("%s: %s \nchanged: %d \n", res.GetTitle().Value, res.GetDetail().Value, res.GetChanged().Value)
	return nil
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

		cwd, err := os.Getwd()
		if err != nil {
			Log.Fatal(err)
		}

		viper.AddConfigPath("/etc/grendel/")
		viper.AddConfigPath(home)
		viper.AddConfigPath(cwd)
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

	err := config.ParseConfigs()
	if err != nil {
		Log.Fatal(err)
	}
}
