// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/config"
	"github.com/ubccr/grendel/pkg/model"
)

type ignitionConfig struct {
	Passwd struct {
		Users []struct {
			Name              string   `json:"name"`
			PasswordHash      string   `json:"passwordHash"`
			SSHAuthorizedKeys []string `json:"sshAuthorizedKeys"`
		} `json:"users"`
	} `json:"passwd"`
	Storage struct {
		Files []struct {
			Path string `json:"path"`
		} `json:"files"`
	} `json:"storage"`
}

// requestIgnition requests the ignition config for a provisioning host using the named template, or the built-in template when empty
func requestIgnition(t *testing.T, tmplName string) (*httptest.ResponseRecorder, error) {
	t.Helper()

	origDNS := config.DefaultDNS
	t.Cleanup(func() { config.DefaultDNS = origDNS })
	config.DefaultDNS = []net.IP{net.ParseIP("10.0.0.2")}

	h := &Handler{DB: newTestDB(t)}

	image := &model.BootImage{Name: "flatcar", KernelPath: "/vmlinuz"}
	if tmplName != "" {
		image.ProvisionTemplates = map[string]string{"butane": tmplName}
	}
	require.NoError(t, h.DB.StoreBootImage(image))

	host := &model.Host{
		Name:       "cpn-01",
		BootImage:  image.Name,
		Provision:  true,
		Interfaces: []*model.NetInterface{{MAC: net.HardwareAddr{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}}},
	}
	require.NoError(t, h.DB.StoreHost(host))

	token, err := model.NewBootToken(host.UID.String(), host.Interfaces[0].MAC.String())
	require.NoError(t, err)

	e := newTestEcho(t)
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
	c.SetPath("/boot/:token/pxe-config.ign")
	c.SetParamNames("token")
	c.SetParamValues(token)

	return rec, TokenRequired(h.Ignition)(c)
}

func TestIgnitionBuiltinTemplate(t *testing.T) {
	tests := []struct {
		name      string
		extraVars map[string]any
		users     map[string]bool
	}{
		{name: "no extra vars", users: map[string]bool{}},
		{name: "root password", extraVars: map[string]any{"root_password": "$6$salt$hash"}, users: map[string]bool{"root": true}},
		{name: "ssh keys", extraVars: map[string]any{"ssh_keys": []string{"ssh-ed25519 AAAA test"}}, users: map[string]bool{"core": true}},
		{name: "root password and ssh keys", extraVars: map[string]any{"root_password": "$6$salt$hash", "ssh_keys": []string{"ssh-ed25519 AAAA test"}}, users: map[string]bool{"root": true, "core": true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("provision.extra_vars", tt.extraVars)
			t.Cleanup(func() { viper.Set("provision.extra_vars", nil) })

			rec, err := requestIgnition(t, "")
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			var ign ignitionConfig
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &ign))

			require.Len(t, ign.Storage.Files, 1)
			assert.Equal(t, "/etc/systemd/network/10-static.network", ign.Storage.Files[0].Path)

			users := map[string]bool{}
			for _, u := range ign.Passwd.Users {
				users[u.Name] = true
				if u.Name == "root" {
					assert.Equal(t, "$6$salt$hash", u.PasswordHash)
				}
				if u.Name == "core" {
					assert.Equal(t, []string{"ssh-ed25519 AAAA test"}, u.SSHAuthorizedKeys)
				}
			}
			assert.Equal(t, tt.users, users)
		})
	}
}

func TestIgnitionStrict(t *testing.T) {
	tests := []struct {
		name     string
		template string
		errMsg   string
	}{
		{
			name:     "valid",
			template: "variant: flatcar\nversion: 1.0.0\nstorage:\n  files:\n    - path: /etc/test\n",
		},
		{
			name:     "misspelled key is rejected",
			template: "variant: flatcar\nversion: 1.0.0\nstorge:\n  files:\n    - path: /etc/test\n",
			errMsg:   "unused key storge",
		},
		{
			name:     "mis-indented key is rejected",
			template: "variant: flatcar\nversion: 1.0.0\npasswd:\n  files:\n    - path: /etc/test\n",
			errMsg:   "unused key files",
		},
		{
			name:     "invalid yaml is rejected",
			template: "variant: flatcar\nversion: 1.0.0\npasswd:\n    - name: root\n  files:\n",
			errMsg:   "did not find expected key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "custom.tmpl"), []byte(tt.template), 0o644))
			viper.Set("provision.templates_dir", dir)
			t.Cleanup(func() { viper.Set("provision.templates_dir", nil) })

			rec, err := requestIgnition(t, "custom.tmpl")

			if tt.errMsg == "" {
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)
				return
			}

			require.Error(t, err)
			he, ok := err.(*echo.HTTPError)
			require.True(t, ok, "expected an echo.HTTPError, got %T", err)
			assert.Equal(t, http.StatusInternalServerError, he.Code)
			assert.ErrorContains(t, he.Internal, "custom.tmpl")
			assert.ErrorContains(t, he.Internal, tt.errMsg)
			assert.Empty(t, rec.Body.String())
		})
	}
}
