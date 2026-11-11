// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/schemas"
)

var logoutTimeout = 10 * time.Second

type conn struct {
	client  *gofish.APIClient
	service *gofish.Service
	target  Target
}

// Close ends the Redfish session. Failed logouts are logged
func (c *conn) Close() {
	if c.client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), logoutTimeout)
	defer cancel()

	client := c.client.WithContext(ctx)

	_, noSession := client.GetSession()
	client.Logout()
	if _, err := client.GetSession(); noSession == nil && err == nil {
		log.Warnf("failed to end the redfish session on %s, it stays open until the bmc expires it", c.target.Name)
	}
}

// attempt login to BMC trying each configured credential
func (s *Service) connect(ctx context.Context, t Target) (*conn, error) {
	if t.Addr == "" {
		return nil, errors.New("bmc interface has no usable ip address")
	}
	if len(s.cfg.Credentials) == 0 {
		return nil, errors.New("no bmc credentials configured, set bmc.user and bmc.password")
	}

	base := s.clientConfig(t)

	var lastErr error
	for i, cred := range s.cfg.Credentials {
		cfg := base
		cfg.Username = cred.User
		cfg.Password = cred.Password

		client, err := gofish.ConnectContext(ctx, cfg)
		if err == nil && s.cfg.BasicAuth && i < len(s.cfg.Credentials)-1 {
			err = checkBasicAuth(client)
		}
		if err == nil {
			return &conn{client: client, service: client.GetService(), target: t}, nil
		}

		lastErr = err
		if !isAuthFailure(err) {
			return nil, err
		}

		log.Debugf("credentials for user %s rejected by %s, trying the next pair", cred.User, t.Name)
	}

	return nil, fmt.Errorf("no configured credentials were accepted (tried %s): %s", strings.Join(s.credentialUsers(), ", "), errorMessage(lastErr))
}

// checkBasicAuth makes the BMC judge the credentials. Connecting with basic auth only stores them, and the service root it fetches is public, so without this every password looks accepted and the next credential is never tried.
func checkBasicAuth(client *gofish.APIClient) error {
	_, err := client.GetService().Systems()
	if isAuthFailure(err) {
		client.Logout()
		return err
	}
	return nil
}

func (s *Service) credentialUsers() []string {
	users := make([]string, 0, len(s.cfg.Credentials))
	for _, c := range s.cfg.Credentials {
		users = append(users, c.User)
	}
	return users
}

// isAuthFailure decides whether to try the next credential, the one place behaviour depends on the kind of error
func isAuthFailure(err error) bool {
	// a collection error holds the real failure per member, so look through it the same way errorMessage does
	redfishErr, ok := errors.AsType[*schemas.Error](parseCollectionErrors(err))
	if !ok {
		return false
	}

	return redfishErr.HTTPReturnedStatusCode == http.StatusUnauthorized || redfishErr.HTTPReturnedStatusCode == http.StatusForbidden
}

func (s *Service) clientConfig(t Target) gofish.ClientConfig {
	return gofish.ClientConfig{
		Endpoint: endpoint(s.cfg.Scheme, t.Addr),
		Insecure: s.cfg.Insecure,
		// BasicAuth avoids creating and tearing down a Redfish session, which is two extra round trips per node per operation
		BasicAuth: s.cfg.BasicAuth,
		HTTPClient: &http.Client{
			Timeout: s.cfg.ClientTimeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   s.cfg.DialTimeout,
					KeepAlive: s.cfg.DialKeepAlive,
				}).DialContext,
				TLSHandshakeTimeout: s.cfg.TLSHandshakeTimeout,
				IdleConnTimeout:     s.cfg.IdleConnTimeout,
			},
		},
		MaxConcurrentRequests: s.cfg.MaxConcurrentRequests,
		ReuseConnections:      s.cfg.ReuseConnections,
	}
}

// endpoint accepts a bare address, an address with a port, or a full url, so tests can point at an httptest server. A bare address gets the configured scheme, which in production is every node, since a bmc interface records an ip and nothing more
func endpoint(scheme, addr string) string {
	if strings.HasPrefix(addr, schemeHTTP+"://") || strings.HasPrefix(addr, schemeHTTPS+"://") {
		return addr
	}
	if scheme == "" {
		scheme = schemeHTTPS
	}
	return scheme + "://" + addr
}
