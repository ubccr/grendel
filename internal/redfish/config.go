// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"github.com/ubccr/grendel/internal/settings"
	"golang.org/x/time/rate"
)

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

type Credential struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type Config struct {
	Credentials []Credential
	Insecure    bool
	BasicAuth   bool
	Scheme      string

	Fanout          int
	NodeTimeout     time.Duration
	RunTimeout      time.Duration
	RateLimitPerSec float64

	ClientTimeout         time.Duration
	DialTimeout           time.Duration
	DialKeepAlive         time.Duration
	TLSHandshakeTimeout   time.Duration
	IdleConnTimeout       time.Duration
	MaxConcurrentRequests int64
	ReuseConnections      bool
	AutoExpand            bool

	IPMIPort         int
	IPMICipherSuites []int
}

func ConfigFromSettings() Config {
	cfg := Config{
		Insecure:  settings.BMCInsecure.Get(),
		BasicAuth: settings.BMCGofishBasicAuth.Get(),
		Scheme:    scheme(settings.BMCScheme.Get()),

		Fanout:          settings.BMCFanout.Get(),
		NodeTimeout:     settings.BMCNodeTimeout.Get(),
		RunTimeout:      settings.BMCRunTimeout.Get(),
		RateLimitPerSec: settings.BMCRateLimitPerSecond.Get(),

		ClientTimeout:       settings.BMCGofishClientTimeout.Get(),
		DialTimeout:         settings.BMCGofishDialTimeout.Get(),
		DialKeepAlive:       settings.BMCGofishDialKeepAliveTimeout.Get(),
		TLSHandshakeTimeout: settings.BMCGofishTLSHandshakeTimeout.Get(),
		IdleConnTimeout:     settings.BMCGofishIdleConnTimeout.Get(),

		MaxConcurrentRequests: int64(settings.BMCGofishMaxConcurrentRequests.Get()),
		ReuseConnections:      settings.BMCGofishReuseConnections.Get(),
		AutoExpand:            settings.BMCGofishAutoExpand.Get(),

		IPMIPort:         settings.BMCIPMIPort.Get(),
		IPMICipherSuites: settings.BMCIPMICipherSuites.Get(),
	}

	if user := cmp.Or(settings.BMCUsername.Get(), settings.BMCUser.Get()); user != "" {
		cfg.Credentials = append(cfg.Credentials, Credential{User: user, Password: settings.BMCPassword.Get()})
	}

	// an array of tables, rather than a username = "password" table, because viper lowercases keys and BMC usernames can be case sensitive. It also keeps the order they are tried in
	for _, row := range slices.Concat(settings.BMCCredentials.Get(), settings.BMCDefaultCredentials.Get()) {
		// the older bmc.defaults.credentials calls the field user
		user, _ := cmp.Or(row["username"], row["user"]).(string)
		password, _ := row["password"].(string)
		if user == "" {
			log.Warn("ignoring a bmc.credentials entry with no username")
			continue
		}
		cfg.Credentials = append(cfg.Credentials, Credential{User: user, Password: password})
	}

	return cfg
}

func scheme(configured string) string {
	switch strings.ToLower(strings.TrimSpace(configured)) {
	case "", schemeHTTPS:
		return schemeHTTPS
	case schemeHTTP:
		log.Warn("bmc.scheme is http, bmc credentials will be sent in plaintext. This is meant for a development redfish mockup server")
		return schemeHTTP
	default:
		log.Warnf("bmc.scheme %q is not http or https, using https", configured)
		return schemeHTTPS
	}
}

func (c Config) options() Options {
	opts := Options{Fanout: c.Fanout, NodeTimeout: c.NodeTimeout}
	if c.RateLimitPerSec > 0 {
		opts.RateLimit = rate.NewLimiter(rate.Limit(c.RateLimitPerSec), 1)
	}
	return opts
}
