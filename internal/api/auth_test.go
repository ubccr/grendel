// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"encoding/json"
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

// signup is unauthenticated, so it must never change an account that already exists, or anyone could set any user's password, the administrator's included
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

func send(t *testing.T, mux http.Handler, method, path, body string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestVerifyUserReportsBadCredentialsWithoutError(t *testing.T) {
	db, err := sqlstore.New(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	_, err = db.StoreUser("alice", "password-1")
	require.NoError(t, err)

	for _, c := range [][2]string{{"alice", "wrong-password"}, {"nobody", "password-1"}} {
		ok, role, err := db.VerifyUser(c[0], c[1])
		assert.NoError(t, err, c)
		assert.False(t, ok, c)
		assert.Empty(t, role, c)
	}
}

// a fault in the store is not bad credentials, so it must not tell the person their password is wrong
func TestSigninStoreFaultIs500(t *testing.T) {
	db, mux := newAuthTestServer(t)
	require.Equal(t, http.StatusOK, postJSON(t, mux, "/v1/auth/signup", `{"username":"alice","password":"password-1"}`).Code)
	require.NoError(t, db.Close())

	rec := postJSON(t, mux, "/v1/auth/signin", `{"username":"alice","password":"password-1"}`)
	assert.Equal(t, http.StatusInternalServerError, rec.Code, rec.Body.String())
}

func TestResetWrongCurrentPasswordIs401(t *testing.T) {
	db, mux := newAuthTestServer(t)
	signup := postJSON(t, mux, "/v1/auth/signup", `{"username":"alice","password":"password-1"}`)
	require.Equal(t, http.StatusOK, signup.Code)
	session := signup.Result().Cookies()

	rec := send(t, mux, http.MethodPatch, "/v1/auth/reset", `{"current_password":"wrong-password","new_password":"password-2"}`, session)
	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
	ok, _, _ := db.VerifyUser("alice", "password-1")
	assert.True(t, ok, "a refused reset keeps the old password")

	rec = send(t, mux, http.MethodPatch, "/v1/auth/reset", `{"current_password":"password-1","new_password":"password-2"}`, session)
	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	ok, _, _ = db.VerifyUser("alice", "password-2")
	assert.True(t, ok)
}

// the user list is readable by more than administrators, so it must not hand out password hashes; the database dump still carries them so a restore keeps passwords
func TestUserListLeavesOutPasswordHashes(t *testing.T) {
	_, mux := newAuthTestServer(t)
	signup := postJSON(t, mux, "/v1/auth/signup", `{"username":"alice","password":"password-1"}`)
	require.Equal(t, http.StatusOK, signup.Code)
	session := signup.Result().Cookies()

	list := send(t, mux, http.MethodGet, "/v1/users", "", session)
	require.Equal(t, http.StatusOK, list.Code, list.Body.String())
	assert.Contains(t, list.Body.String(), `"username":"alice"`)
	assert.NotContains(t, list.Body.String(), `"hash"`)
	assert.NotContains(t, list.Body.String(), "$2a$")

	dump := send(t, mux, http.MethodGet, "/v1/db/dump", "", session)
	require.Equal(t, http.StatusOK, dump.Code, dump.Body.String())
	assert.Contains(t, dump.Body.String(), `"hash":"$2a$`)
}

func signinAs(t *testing.T, db *sqlstore.SqlStore, mux http.Handler, username, role string) []*http.Cookie {
	t.Helper()
	require.Equal(t, http.StatusOK, postJSON(t, mux, "/v1/auth/signup", `{"username":"`+username+`","password":"password-1"}`).Code)
	require.NoError(t, db.UpdateUserRole(username, role))
	require.NoError(t, db.UpdateUserEnabled(username, true))
	signin := postJSON(t, mux, "/v1/auth/signin", `{"username":"`+username+`","password":"password-1"}`)
	require.Equal(t, http.StatusOK, signin.Code, signin.Body.String())
	return signin.Result().Cookies()
}

// the user role may tag nodes and list roles; read-only may only list roles
func TestUserRoleMayTagNodesAndListRoles(t *testing.T) {
	db, mux := newAuthTestServer(t)
	signinAs(t, db, mux, "admin", model.RoleAdmin.String())
	user := signinAs(t, db, mux, "alice", model.RoleUser.String())
	readOnly := signinAs(t, db, mux, "bob", model.RoleReadOnly.String())

	tag := func(session []*http.Cookie) int {
		return send(t, mux, http.MethodPatch, "/v1/nodes/tags/add?nodeset=cpn-01", `{"tags":"test"}`, session).Code
	}
	roles := func(session []*http.Cookie) int {
		return send(t, mux, http.MethodGet, "/v1/roles", "", session).Code
	}

	assert.NotEqual(t, http.StatusForbidden, tag(user))
	assert.Equal(t, http.StatusOK, roles(user))
	assert.Equal(t, http.StatusForbidden, tag(readOnly))
	assert.Equal(t, http.StatusOK, roles(readOnly))
}

func TestUserRoleMayNotCreateUsers(t *testing.T) {
	db, mux := newAuthTestServer(t)
	admin := signinAs(t, db, mux, "admin", model.RoleAdmin.String())
	user := signinAs(t, db, mux, "alice", model.RoleUser.String())

	create := func(session []*http.Cookie, name string) int {
		return send(t, mux, http.MethodPost, "/v1/users", `{"username":"`+name+`","password":"password-1"}`, session).Code
	}

	assert.Equal(t, http.StatusForbidden, create(user, "carol"))
	assert.Equal(t, http.StatusOK, create(admin, "dave"))
}

func TestRolesDescribeTheirPermissions(t *testing.T) {
	db, mux := newAuthTestServer(t)
	admin := signinAs(t, db, mux, "admin", model.RoleAdmin.String())

	res := send(t, mux, http.MethodGet, "/v1/roles?name=admin", "", admin)
	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	var body GetRolesResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
	require.Len(t, body.Roles, 1)

	byKey := map[string]model.Permission{}
	for _, p := range body.Roles[0].PermissionList {
		byKey[p.Method+" "+p.Path] = p
		assert.NotEmpty(t, p.Summary, "%s %s has no summary", p.Method, p.Path)
		assert.NotEmpty(t, p.Group, "%s %s has no group", p.Method, p.Path)
		assert.NotEmpty(t, p.Kind, "%s %s has no kind", p.Method, p.Path)
	}
	assert.Equal(t, model.Permission{Method: "POST", Path: "/v1/bmc/sel/clear", Summary: "Clear the system event log", Group: "BMC", Kind: model.PermissionDelete}, byKey["POST /v1/bmc/sel/clear"])
	assert.Equal(t, model.Permission{Method: "PATCH", Path: "/v1/nodes/tags/%", Summary: "Add and remove tags", Group: "Nodes", Kind: model.PermissionChange}, byKey["PATCH /v1/nodes/tags/%"])
	assert.Equal(t, model.Permission{Method: "GET", Path: "/v1/bmc/console", Summary: "Open the serial console", Group: "BMC", Kind: model.PermissionChange}, byKey["GET /v1/bmc/console"])
}
