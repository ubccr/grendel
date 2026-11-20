// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"cmp"
	"time"

	"github.com/spf13/viper"
)

type GNMIConfig struct {
	Port            int    `mapstructure:"port"`
	Plaintext       bool   `mapstructure:"plaintext"`
	SkipVerify      bool   `mapstructure:"skip_verify"`
	CACert          string `mapstructure:"ca_cert"`
	Target          string `mapstructure:"target"`
	NetworkInstance string `mapstructure:"network_instance"`
}

type SNMPConfig struct {
	Version   string        `mapstructure:"version"`
	Community string        `mapstructure:"community"`
	Port      int           `mapstructure:"port"`
	Timeout   time.Duration `mapstructure:"timeout"`
	Retries   int           `mapstructure:"retries"`

	User      string `mapstructure:"user"`
	AuthProto string `mapstructure:"auth_proto"`
	AuthPass  string `mapstructure:"auth_pass"`
	PrivProto string `mapstructure:"priv_proto"`
	PrivPass  string `mapstructure:"priv_pass"`
}

type Config struct {
	Username string        `mapstructure:"username"`
	Password string        `mapstructure:"password"`
	Timeout  time.Duration `mapstructure:"timeout"`
	GNMI     GNMIConfig    `mapstructure:"gnmi"`
	SNMP     SNMPConfig    `mapstructure:"snmp"`

	gnmiByNOS map[string]GNMIConfig
}

// defaultConfig is decoded over rather than set with viper.SetDefault, since viper leaves nested defaults out of a table once any key in it is set
var defaultConfig = Config{
	Timeout: 30 * time.Second,
	GNMI:    GNMIConfig{Port: 6030, SkipVerify: true, NetworkInstance: "default"},
	SNMP:    SNMPConfig{Version: "2c", Community: "public", Port: 161, Timeout: 5 * time.Second, Retries: 2},
}

func ConfigFromViper() Config {
	cfg := defaultConfig
	if err := viper.UnmarshalKey("switch", &cfg); err != nil {
		log.Warnf("reading the [switch] config: %v", err)
	}

	// a [switch.gnmi.<nos>] table overrides only the keys it sets, since decoding into a copy keeps the fields the table leaves out
	cfg.gnmiByNOS = map[string]GNMIConfig{}
	for _, nos := range []string{NOSArista, NOSSonic} {
		g := cfg.GNMI
		if sub := viper.Sub("switch.gnmi." + nos); sub != nil {
			if err := sub.Unmarshal(&g); err != nil {
				log.Warnf("reading the [switch.gnmi.%s] config: %v", nos, err)
			}
		}
		cfg.gnmiByNOS[nos] = g
	}

	// each falls back on its own, so a config that sets only switch.username still finds the password under the older key
	cfg.Username = cmp.Or(cfg.Username, viper.GetString("bmc.switch_admin_username"))
	cfg.Password = cmp.Or(cfg.Password, viper.GetString("bmc.switch_admin_password"))

	return cfg
}

func (c Config) gnmiFor(nos string) GNMIConfig {
	if g, ok := c.gnmiByNOS[nos]; ok {
		return g
	}
	return c.GNMI
}
