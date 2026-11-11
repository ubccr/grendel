// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"strings"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

func init() {
	viper.SetDefault("bmc.fanout", 5)

	viper.SetDefault("bmc.scheme", schemeHTTPS)

	viper.SetDefault("bmc.node_timeout", time.Second*60)
	viper.SetDefault("bmc.run_timeout", time.Minute*4)
	viper.SetDefault("bmc.rate_limit_per_second", 0)

	viper.SetDefault("bmc.gofish.basic_auth", false)
	viper.SetDefault("bmc.gofish.max_concurrent_requests", 1)
	viper.SetDefault("bmc.gofish.reuse_connections", false)
	viper.SetDefault("bmc.gofish.client_timeout", time.Second*60)
	viper.SetDefault("bmc.gofish.dial_timeout", time.Second*5)
	viper.SetDefault("bmc.gofish.dial_keep_alive_timeout", time.Second*30)
	viper.SetDefault("bmc.gofish.tls_handshake_timeout", time.Second*10)
	viper.SetDefault("bmc.gofish.idle_conn_timeout", time.Second*90)

	viper.SetDefault("bmc.ipmi.port", 623)
	viper.SetDefault("bmc.ipmi.cipher_suites", []int{17, 3})
}

// BMC username and password
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

	IPMIPort         int
	IPMICipherSuites []int
}

func ConfigFromViper() Config {
	cfg := Config{
		Insecure:  viper.GetBool("bmc.insecure"),
		BasicAuth: viper.GetBool("bmc.gofish.basic_auth"),
		Scheme:    scheme(viper.GetString("bmc.scheme")),

		Fanout:          viper.GetInt("bmc.fanout"),
		NodeTimeout:     viper.GetDuration("bmc.node_timeout"),
		RunTimeout:      viper.GetDuration("bmc.run_timeout"),
		RateLimitPerSec: viper.GetFloat64("bmc.rate_limit_per_second"),

		ClientTimeout:       viper.GetDuration("bmc.gofish.client_timeout"),
		DialTimeout:         viper.GetDuration("bmc.gofish.dial_timeout"),
		DialKeepAlive:       viper.GetDuration("bmc.gofish.dial_keep_alive_timeout"),
		TLSHandshakeTimeout: viper.GetDuration("bmc.gofish.tls_handshake_timeout"),
		IdleConnTimeout:     viper.GetDuration("bmc.gofish.idle_conn_timeout"),

		MaxConcurrentRequests: viper.GetInt64("bmc.gofish.max_concurrent_requests"),
		ReuseConnections:      viper.GetBool("bmc.gofish.reuse_connections"),

		IPMIPort:         viper.GetInt("bmc.ipmi.port"),
		IPMICipherSuites: viper.GetIntSlice("bmc.ipmi.cipher_suites"),
	}

	if user := viper.GetString("bmc.user"); user != "" {
		cfg.Credentials = append(cfg.Credentials, Credential{User: user, Password: viper.GetString("bmc.password")})
	}

	// an array of tables, rather than a username = "password" table, because viper lowercases keys and BMC usernames can be case sensitive. It also keeps the order they are tried in
	var defaults []Credential
	if err := viper.UnmarshalKey("bmc.defaults.credentials", &defaults); err != nil {
		log.Warnf("ignoring bmc.defaults.credentials, it must be a list of [[bmc.defaults.credentials]] tables with user and password: %s", err)
	}
	for _, c := range defaults {
		if c.User == "" {
			log.Warn("ignoring a bmc.defaults.credentials entry with no user")
			continue
		}
		cfg.Credentials = append(cfg.Credentials, c)
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
