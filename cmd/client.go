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
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/ubccr/grendel/pkg/client"
)

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

	httpClient := &http.Client{Timeout: time.Second * 3600, Transport: tr}
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

	return fmt.Errorf("API Error: status=%d title=%s detail=%s", t.StatusCode, httpError.GetTitle().Value, httpError.GetDetail().Value)
}

func NewApiResponse(res *client.GenericResponse) error {
	fmt.Printf("%s: %s \nchanged: %d \n", res.GetTitle().Value, res.GetDetail().Value, res.GetChanged().Value)
	return nil
}
