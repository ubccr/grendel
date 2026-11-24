// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package tors

import (
	"cmp"
	"time"

	"github.com/ubccr/grendel/internal/settings"
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

func ConfigFromSettings() Config {
	cfg := Config{
		// each falls back on its own, so a config that sets only switch.username still finds the password under the older key
		Username: cmp.Or(settings.SwitchUsername.Get(), settings.BMCSwitchAdminUsername.Get()),
		Password: cmp.Or(settings.SwitchPassword.Get(), settings.BMCSwitchAdminPassword.Get()),
		Timeout:  settings.SwitchTimeout.Get(),
		GNMI:     gnmiConfig(settings.SwitchGNMI),
		SNMP: SNMPConfig{
			Version:   settings.SwitchSNMPVersion.Get(),
			Community: settings.SwitchSNMPCommunity.Get(),
			Port:      settings.SwitchSNMPPort.Get(),
			Timeout:   settings.SwitchSNMPTimeout.Get(),
			Retries:   settings.SwitchSNMPRetries.Get(),
			User:      cmp.Or(settings.SwitchSNMPUsername.Get(), settings.SwitchSNMPUser.Get()),
			AuthProto: settings.SwitchSNMPAuthProto.Get(),
			AuthPass:  settings.SwitchSNMPAuthPass.Get(),
			PrivProto: settings.SwitchSNMPPrivProto.Get(),
			PrivPass:  settings.SwitchSNMPPrivPass.Get(),
		},
	}

	cfg.gnmiByNOS = map[string]GNMIConfig{
		NOSArista: gnmiConfig(settings.SwitchGNMIArista),
		NOSSonic:  gnmiConfig(settings.SwitchGNMISonic),
	}

	return cfg
}

func gnmiConfig(o settings.GNMIOptions) GNMIConfig {
	return GNMIConfig{
		Port:            o.Port.Get(),
		Plaintext:       o.Plaintext.Get(),
		SkipVerify:      o.SkipVerify.Get(),
		CACert:          o.CACert.Get(),
		Target:          o.Target.Get(),
		NetworkInstance: o.NetworkInstance.Get(),
	}
}

func (c Config) gnmiFor(nos string) GNMIConfig {
	if g, ok := c.gnmiByNOS[nos]; ok {
		return g
	}
	return c.GNMI
}
