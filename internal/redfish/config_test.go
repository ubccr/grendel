// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/internal/settings/settingstest"
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

func TestConfigFromSettingsScheme(t *testing.T) {
	assert.Equal(t, schemeHTTPS, ConfigFromSettings().Scheme, "the default is https, without any config file")

	settingstest.Set(t, "bmc.scheme", schemeHTTP)

	assert.Equal(t, schemeHTTP, ConfigFromSettings().Scheme)
}

// only what is configured is ever sent to a bmc, so with nothing set there is no login to try
func TestConfigFromSettingsCredentials(t *testing.T) {
	configured := Credential{User: "ops", Password: "secret"}

	tests := []struct {
		name string
		user string
		rows []map[string]any
		want []Credential
	}{
		{
			name: "nothing configured means no credentials",
		},
		{
			name: "bmc.user alone adds no default login",
			user: "ops",
			want: []Credential{configured},
		},
		{
			name: "defaults set in the config are tried after bmc.user",
			user: "ops", rows: []map[string]any{{"user": "root", "password": "calvin"}},
			want: []Credential{configured, {User: "root", Password: "calvin"}},
		},
		{
			name: "defaults work without bmc.user",
			rows: []map[string]any{{"user": "root", "password": "calvin"}},
			want: []Credential{{User: "root", Password: "calvin"}},
		},
		{
			name: "an empty list adds nothing",
			user: "ops", rows: []map[string]any{},
			want: []Credential{configured},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.user != "" {
				settingstest.Set(t, "bmc.user", tt.user)
				settingstest.Set(t, "bmc.password", "secret")
			}
			if tt.rows != nil {
				settingstest.Set(t, "bmc.defaults.credentials", tt.rows)
			}
			assert.Equal(t, tt.want, ConfigFromSettings().Credentials)
		})
	}
}
