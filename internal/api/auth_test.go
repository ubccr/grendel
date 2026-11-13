// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-fuego/fuego"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/sqlstore"
	"github.com/ubccr/grendel/pkg/model"
)

// newAuthTestServer serves the real routes over an in-memory store, wired the way Serve wires them
func newAuthTestServer(t *testing.T) (*sqlstore.SqlStore, http.Handler) {
	t.Helper()
	viper.Set("api.secret", "auth-test-secret-0123456789abcdef")
	t.Cleanup(func() { viper.Set("api.secret", nil) })

	db, err := sqlstore.New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	h, err := NewHandler(db)
	require.NoError(t, err)

	s := fuego.NewServer(
		fuego.WithEngineOptions(fuego.WithErrorHandler(ErrorHandler)),
		fuego.WithErrorSerializer(ErrorSerializer),
		fuego.WithSecurity(setupSecurity()),
		fuego.WithoutStartupMessages(),
	)
	h.SetupRoutes(s)
	return db, s.Mux
}

func postJSON(t *testing.T, mux http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// signup is unauthenticated, so it must never change an account that already exists; it used to upsert, which let anyone set any user's password, the administrator's included
func TestSignupCannotTakeOverAnExistingUser(t *testing.T) {
	db, mux := newAuthTestServer(t)

	first := postJSON(t, mux, "/v1/auth/signup", `{"username":"admin","password":"original-password"}`)
	require.Equal(t, http.StatusOK, first.Code, first.Body.String())
	assert.Contains(t, first.Body.String(), `"role":"admin"`)

	again := postJSON(t, mux, "/v1/auth/signup", `{"username":"admin","password":"attacker-password"}`)
	assert.Equal(t, http.StatusConflict, again.Code, again.Body.String())
	assert.Contains(t, again.Body.String(), "username already taken")
	assert.Contains(t, again.Body.String(), `"status":409`)
	assert.Empty(t, again.Result().Cookies(), "a refused signup must not hand out a session")

	ok, role, err := db.VerifyUser("admin", "original-password")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, model.RoleAdmin.String(), role)

	ok, _, _ = db.VerifyUser("admin", "attacker-password")
	assert.False(t, ok)

	signin := postJSON(t, mux, "/v1/auth/signin", `{"username":"admin","password":"attacker-password"}`)
	assert.Equal(t, http.StatusUnauthorized, signin.Code)
}

// the web UI tells bad credentials apart from a server fault by status, so both a wrong password and an unknown user must answer 401 with the status in the body
func TestSigninBadCredentialsIs401(t *testing.T) {
	_, mux := newAuthTestServer(t)
	require.Equal(t, http.StatusOK, postJSON(t, mux, "/v1/auth/signup", `{"username":"alice","password":"password-1"}`).Code)

	for _, body := range []string{`{"username":"alice","password":"wrong-password"}`, `{"username":"nobody","password":"password-1"}`} {
		rec := postJSON(t, mux, "/v1/auth/signin", body)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, body)
		assert.Contains(t, rec.Body.String(), `"status":401`, body)
	}
}

func TestStoreUserRefusesDuplicate(t *testing.T) {
	db, err := sqlstore.New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.StoreUser("alice", "password-1")
	require.NoError(t, err)

	_, err = db.StoreUser("alice", "password-2")
	assert.ErrorIs(t, err, store.ErrDuplicateEntry)
}

func TestUpdateUserPassword(t *testing.T) {
	db, err := sqlstore.New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.StoreUser("alice", "password-1")
	require.NoError(t, err)

	require.NoError(t, db.UpdateUserPassword("alice", "password-2"))
	ok, role, err := db.VerifyUser("alice", "password-2")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, model.RoleAdmin.String(), role, "a password change keeps the role")

	assert.ErrorIs(t, db.UpdateUserPassword("nobody", "password-3"), store.ErrNotFound)
}
