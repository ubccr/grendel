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
		old  []map[string]any
		want []Credential
	}{
		{
			name: "nothing configured means no credentials",
		},
		{
			name: "bmc.username alone is the only login",
			user: "ops",
			want: []Credential{configured},
		},
		{
			name: "the list is tried after bmc.username",
			user: "ops", rows: []map[string]any{{"username": "root", "password": "calvin"}},
			want: []Credential{configured, {User: "root", Password: "calvin"}},
		},
		{
			name: "the list works without bmc.username",
			rows: []map[string]any{{"username": "root", "password": "calvin"}},
			want: []Credential{{User: "root", Password: "calvin"}},
		},
		{
			name: "logins under the older key are tried last",
			rows: []map[string]any{{"username": "root", "password": "calvin"}},
			old:  []map[string]any{{"user": "ADMIN", "password": "ADMIN"}},
			want: []Credential{{User: "root", Password: "calvin"}, {User: "ADMIN", Password: "ADMIN"}},
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
				settingstest.Set(t, "bmc.username", tt.user)
				settingstest.Set(t, "bmc.password", "secret")
			}
			if tt.rows != nil {
				settingstest.Set(t, "bmc.credentials", tt.rows)
			}
			if tt.old != nil {
				settingstest.Set(t, "bmc.defaults.credentials", tt.old)
			}
			assert.Equal(t, tt.want, ConfigFromSettings().Credentials)
		})
	}
}
