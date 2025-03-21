// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/sqlstore"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
)

func newTestDB(t *testing.T) store.Store {
	assert := assert.New(t)

	db, err := sqlstore.New(":memory:")
	if err != nil {
		assert.Fail(err.Error())
	}

	return db
}

func newTestEcho(t *testing.T) *echo.Echo {
	e, err := newEcho()
	if err != nil {
		assert.Fail(t, err.Error())
	}

	return e
}

func TestStatus(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{DB: newTestDB(t)}

	e := newTestEcho(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.Index(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal("up", gjson.Get(rec.Body.String(), "status").String())
	}
}

func TestInvalidBootToken(t *testing.T) {
	assert := assert.New(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	token, err := model.NewBootToken(host.UID.String(), host.Interfaces[0].MAC.String())
	assert.NoError(err)
	badToken := []byte(token)
	badToken[2] = 'a'

	badData := []string{
		"bad token",
		string(badToken),
	}

	h := &Handler{DB: newTestDB(t)}

	for _, test := range badData {
		e := newTestEcho(t)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/boot/:token/ipxe")
		c.SetParamNames("token")
		c.SetParamValues(test)

		err := TokenRequired(h.Ipxe)(c)
		if assert.Error(err) {
			he, ok := err.(*echo.HTTPError)
			if ok {
				assert.Equal(http.StatusBadRequest, he.Code)
			}
			e.HTTPErrorHandler(err, c)
			assert.Equal(http.StatusBadRequest, rec.Code)
			assert.Equal("invalid token", gjson.Get(rec.Body.String(), "message").String())
		}
	}
}

func TestIpxe(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{DB: newTestDB(t)}

	image := tests.BootImageFactory.MustCreate().(*model.BootImage)
	err := h.DB.StoreBootImage(image)
	assert.NoError(err)

	host := tests.HostFactory.MustCreate().(*model.Host)
	host.BootImage = image.Name
	host.Provision = true
	err = h.DB.StoreHost(host)
	assert.NoError(err)

	token, err := model.NewBootToken(host.UID.String(), host.Interfaces[0].MAC.String())
	assert.NoError(err)

	e := newTestEcho(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/boot/:token/ipxe")
	c.SetParamNames("token")
	c.SetParamValues(token)

	if assert.NoError(TokenRequired(h.Ipxe)(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Contains(rec.Body.String(), "#!ipxe")
	}
}

func TestHostNotProvision(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{DB: newTestDB(t)}

	paths := map[string]echo.HandlerFunc{
		"ipxe":      h.Ipxe,
		"kickstart": h.Kickstart,
	}

	image := tests.BootImageFactory.MustCreate().(*model.BootImage)
	err := h.DB.StoreBootImage(image)
	assert.NoError(err)

	host := tests.HostFactory.MustCreate().(*model.Host)
	host.BootImage = image.Name
	host.Provision = false
	err = h.DB.StoreHost(host)
	assert.NoError(err)

	token, err := model.NewBootToken(host.UID.String(), host.Interfaces[0].MAC.String())
	assert.NoError(err)

	for path, handler := range paths {
		e := newTestEcho(t)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath(fmt.Sprintf("/boot/:token/%s", path))
		c.SetParamNames("token")
		c.SetParamValues(token)

		err = TokenRequired(handler)(c)
		if assert.Errorf(err, "no error for %s", path) {
			he, ok := err.(*echo.HTTPError)
			if ok {
				assert.Equalf(http.StatusBadRequest, he.Code, "bad http error code for %s", path)
			}
			e.HTTPErrorHandler(err, c)
			assert.Equalf(http.StatusBadRequest, rec.Code, "bad error code for %s", path)
			assert.Equalf("host not set to provision", gjson.Get(rec.Body.String(), "message").String(), "bad message for %s", path)
		}
	}
}

func TestIpxeWrongHost(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{DB: newTestDB(t)}

	image := tests.BootImageFactory.MustCreate().(*model.BootImage)
	err := h.DB.StoreBootImage(image)
	assert.NoError(err)

	host := tests.HostFactory.MustCreate().(*model.Host)
	host.BootImage = image.Name
	host.Provision = true
	err = h.DB.StoreHost(host)
	assert.NoError(err)

	hostBad := tests.HostFactory.MustCreate().(*model.Host)

	token, err := model.NewBootToken(hostBad.UID.String(), hostBad.Interfaces[0].MAC.String())
	assert.NoError(err)

	e := newTestEcho(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/boot/:token/ipxe")
	c.SetParamNames("token")
	c.SetParamValues(token)

	err = TokenRequired(h.Ipxe)(c)
	if assert.Error(err) {
		he, ok := err.(*echo.HTTPError)
		if ok {
			assert.Equal(http.StatusBadRequest, he.Code)
		}
		e.HTTPErrorHandler(err, c)
		assert.Equal(http.StatusBadRequest, rec.Code)
		assert.Equal("invalid host", gjson.Get(rec.Body.String(), "message").String())
	}
}

func TestKickstart(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{DB: newTestDB(t)}

	image := tests.BootImageFactory.MustCreate().(*model.BootImage)
	err := h.DB.StoreBootImage(image)
	assert.NoError(err)

	host := tests.HostFactory.MustCreate().(*model.Host)
	host.BootImage = image.Name
	host.Provision = true
	err = h.DB.StoreHost(host)
	assert.NoError(err)

	token, err := model.NewBootToken(host.UID.String(), host.Interfaces[0].MAC.String())
	assert.NoError(err)

	e := newTestEcho(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/boot/:token/kickstart")
	c.SetParamNames("token")
	c.SetParamValues(token)

	if assert.NoError(TokenRequired(h.Kickstart)(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Contains(rec.Body.String(), "install")
		assert.Contains(rec.Body.String(), "liveimg --url=")
	}
}

func TestComplete(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{DB: newTestDB(t)}

	image := tests.BootImageFactory.MustCreate().(*model.BootImage)
	err := h.DB.StoreBootImage(image)
	assert.NoError(err)

	host := tests.HostFactory.MustCreate().(*model.Host)
	host.BootImage = image.Name
	host.Provision = true
	err = h.DB.StoreHost(host)
	assert.NoError(err)

	token, err := model.NewBootToken(host.UID.String(), host.Interfaces[0].MAC.String())
	assert.NoError(err)

	e := newTestEcho(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/boot/:token/complete")
	c.SetParamNames("token")
	c.SetParamValues(token)

	if assert.NoError(TokenRequired(h.Complete)(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal("ok", gjson.Get(rec.Body.String(), "status").String())
	}

	hostTest, err := h.DB.LoadHostFromID(host.UID.String())
	if assert.NoError(err) {
		assert.False(hostTest.Provision)
	}
}

func TestUserData(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{DB: newTestDB(t)}

	image := tests.BootImageFactory.MustCreate().(*model.BootImage)
	err := h.DB.StoreBootImage(image)
	assert.NoError(err)

	host := tests.HostFactory.MustCreate().(*model.Host)
	host.BootImage = image.Name
	host.Provision = true
	err = h.DB.StoreHost(host)
	assert.NoError(err)

	token, err := model.NewBootToken(host.UID.String(), host.Interfaces[0].MAC.String())
	assert.NoError(err)

	e := newTestEcho(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/boot/:token/cloud-init/user-data")
	c.SetParamNames("token")
	c.SetParamValues(token)

	if assert.NoError(TokenRequired(h.UserData)(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Contains(rec.Body.String(), "phone_home")
		assert.Contains(rec.Body.String(), "final_message")
	}
}

func TestMetaData(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{DB: newTestDB(t)}

	image := tests.BootImageFactory.MustCreate().(*model.BootImage)
	err := h.DB.StoreBootImage(image)
	assert.NoError(err)

	host := tests.HostFactory.MustCreate().(*model.Host)
	host.BootImage = image.Name
	host.Provision = true
	err = h.DB.StoreHost(host)
	assert.NoError(err)

	token, err := model.NewBootToken(host.UID.String(), host.Interfaces[0].MAC.String())
	assert.NoError(err)

	e := newTestEcho(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/boot/:token/cloud-init/meta-data")
	c.SetParamNames("token")
	c.SetParamValues(token)

	if assert.NoError(TokenRequired(h.MetaData)(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Contains(rec.Body.String(), "instance-id")
		assert.Contains(rec.Body.String(), host.UID.String())
	}
}
