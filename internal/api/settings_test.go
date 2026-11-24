// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/settings"
	"github.com/ubccr/grendel/internal/store/sqlstore"
	"github.com/ubccr/grendel/pkg/model"
)

// stored lists the settings rows other than the secrets grendel generates for itself
func stored(t *testing.T, db *sqlstore.SqlStore) []settings.Row {
	t.Helper()
	rows, _, err := db.Settings()
	require.NoError(t, err)
	return slices.DeleteFunc(rows, func(r settings.Row) bool { return strings.HasSuffix(r.Key, ".secret") })
}

func TestSettingsAreOnlyForAnAdmin(t *testing.T) {
	db, mux := newAuthTestServer(t)
	settings.Reset()
	t.Cleanup(settings.Reset)
	_, err := settings.Load(db)
	require.NoError(t, err)

	admin := signinAs(t, db, mux, "admin", model.RoleAdmin.String())
	user := signinAs(t, db, mux, "alice", model.RoleUser.String())
	readOnly := signinAs(t, db, mux, "bob", model.RoleReadOnly.String())

	list := send(t, mux, http.MethodGet, "/v1/settings", "", admin)
	assert.Equal(t, http.StatusOK, list.Code)
	assert.NotContains(t, list.Body.String(), settings.APISecret.Get())
	assert.NotContains(t, list.Body.String(), settings.ProvisionSecret.Get())
	for _, session := range [][]*http.Cookie{user, readOnly} {
		assert.Equal(t, http.StatusForbidden, send(t, mux, http.MethodGet, "/v1/settings", "", session).Code)
		assert.Equal(t, http.StatusForbidden, send(t, mux, http.MethodPatch, "/v1/settings", `{"set":{"dns.ttl":"600"}}`, session).Code)
	}
}

func TestPatchSettingsStoresAndReportsEachRefusal(t *testing.T) {
	db, mux := newAuthTestServer(t)
	settings.Reset()
	t.Cleanup(settings.Reset)
	_, err := settings.Load(db)
	require.NoError(t, err)
	admin := signinAs(t, db, mux, "admin", model.RoleAdmin.String())

	bad := send(t, mux, http.MethodPatch, "/v1/settings", `{"set":{"dns.forward":"9.9.9.9:53","dns.ttl":"-5","api.secret":"x"}}`, admin)
	require.Equal(t, http.StatusUnprocessableEntity, bad.Code, bad.Body.String())
	assert.Contains(t, bad.Body.String(), `"name":"dns.ttl"`)
	assert.Contains(t, bad.Body.String(), `"name":"api.secret"`)
	assert.Empty(t, stored(t, db), "one refused change must leave nothing saved")

	// a value is taken as text or in the setting's own type, and comes back in its own type
	ok := send(t, mux, http.MethodPatch, "/v1/settings", `{"set":{"dns.forward":"9.9.9.9:53","dns.ttl":"600","dhcp.mtu":9000,"dhcp.dns_servers":["10.0.0.2"]}}`, admin)
	require.Equal(t, http.StatusOK, ok.Code, ok.Body.String())
	assert.Contains(t, ok.Body.String(), `"pending":1`)
	assert.Contains(t, ok.Body.String(), `"value":600`)
	assert.Contains(t, ok.Body.String(), `"value":["10.0.0.2"]`)
	assert.Equal(t, "9.9.9.9:53", settings.DNSForward.Get())
	assert.Equal(t, 9000, settings.DHCPMTU.Get())

	rows := stored(t, db)
	require.Len(t, rows, 4)
	assert.Equal(t, "admin", rows[0].UpdatedBy)

	unset := send(t, mux, http.MethodPatch, "/v1/settings", `{"unset":["dns.forward","dns.ttl","dhcp.mtu","dhcp.dns_servers"]}`, admin)
	require.Equal(t, http.StatusOK, unset.Code, unset.Body.String())
	assert.Empty(t, stored(t, db))
}

// a server with no secret set keeps the one it generates, so sessions survive a restart
func TestGeneratedSecretsSurviveAReload(t *testing.T) {
	db, _ := newAuthTestServer(t)
	settings.Reset()
	t.Cleanup(settings.Reset)
	_, err := settings.Load(db)
	require.NoError(t, err)
	first := settings.ProvisionSecret.Get()
	require.Len(t, first, 32)

	settings.Reset()
	_, err = settings.Load(db)
	require.NoError(t, err)
	assert.Equal(t, first, settings.ProvisionSecret.Get())
}

func TestDumpCarriesSettingsButNotGeneratedSecrets(t *testing.T) {
	db, mux := newAuthTestServer(t)
	settings.Reset()
	t.Cleanup(settings.Reset)
	_, err := settings.Load(db)
	require.NoError(t, err)
	admin := signinAs(t, db, mux, "admin", model.RoleAdmin.String())
	require.Equal(t, http.StatusOK, send(t, mux, http.MethodPatch, "/v1/settings", `{"set":{"dns.ttl":"600"}}`, admin).Code)

	dump := send(t, mux, http.MethodGet, "/v1/db/dump", "", admin)
	require.Equal(t, http.StatusOK, dump.Code, dump.Body.String())
	assert.Contains(t, dump.Body.String(), `"key":"dns.ttl"`)
	assert.NotContains(t, dump.Body.String(), "api.secret")
	assert.NotContains(t, dump.Body.String(), settings.ProvisionSecret.Get())
}
