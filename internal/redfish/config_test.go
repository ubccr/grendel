// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestScheme(t *testing.T) {
	tests := []struct {
		configured string
		want       string
	}{
		{"", schemeHTTPS},
		{"https", schemeHTTPS},
		{"http", schemeHTTP},
		{"HTTP", schemeHTTP},
		{" http ", schemeHTTP},
		// a typo must not quietly downgrade every bmc connection in the cluster
		{"htp", schemeHTTPS},
		{"ftp", schemeHTTPS},
		{"https://", schemeHTTPS},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, scheme(tt.configured), "bmc.scheme = %q", tt.configured)
	}
}

// TestConfigFromViperScheme pins the key name, since viper returns a quiet zero value for a misspelled one
func TestConfigFromViperScheme(t *testing.T) {
	assert.Equal(t, schemeHTTPS, ConfigFromViper().Scheme, "the default is https, without any config file")

	viper.Set("bmc.scheme", schemeHTTP)
	t.Cleanup(func() { viper.Set("bmc.scheme", schemeHTTPS) })

	assert.Equal(t, schemeHTTP, ConfigFromViper().Scheme)
}

// TestConfigFromViperCredentials reads real toml rather than using viper.Set, because the point is that only what a config file sets is ever sent to a bmc
func TestConfigFromViperCredentials(t *testing.T) {
	configured := Credential{User: "ops", Password: "secret"}

	tests := []struct {
		name string
		toml string
		want []Credential
	}{
		{
			name: "nothing configured means no credentials",
		},
		{
			name: "bmc.user alone adds no default login",
			toml: "[bmc]\nuser = \"ops\"\npassword = \"secret\"\n",
			want: []Credential{configured},
		},
		{
			name: "defaults set in the config are tried after bmc.user",
			toml: "[bmc]\nuser = \"ops\"\npassword = \"secret\"\n[[bmc.defaults.credentials]]\nuser = \"root\"\npassword = \"calvin\"\n",
			want: []Credential{configured, {User: "root", Password: "calvin"}},
		},
		{
			name: "defaults work without bmc.user",
			toml: "[[bmc.defaults.credentials]]\nuser = \"root\"\npassword = \"calvin\"\n",
			want: []Credential{{User: "root", Password: "calvin"}},
		},
		{
			name: "an empty list adds nothing",
			toml: "[bmc]\nuser = \"ops\"\npassword = \"secret\"\n[bmc.defaults]\ncredentials = []\n",
			want: []Credential{configured},
		},
	}

	viper.SetConfigType("toml")
	t.Cleanup(func() { assert.NoError(t, viper.ReadConfig(strings.NewReader(""))) })

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, viper.ReadConfig(strings.NewReader(tt.toml)))
			assert.Equal(t, tt.want, ConfigFromViper().Credentials)
		})
	}
}
