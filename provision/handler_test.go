// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package provision

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/model"
)

func newTestDB(t *testing.T) model.DataStore {
	assert := assert.New(t)

	db, err := model.NewDataStore(":memory:")
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

func TestIpxeInvalidToken(t *testing.T) {
	assert := assert.New(t)

	host := tests.HostFactory.MustCreate().(*model.Host)
	token, err := model.NewBootToken(host.ID.String(), host.Interfaces[0].MAC.String())
	assert.NoError(err)
	badToken := []byte(token)
	badToken[2] = 'a'

	badData := []string{
		"bad token",
		string(badToken),
	}

	h := &Handler{DB: newTestDB(t)}

	for _, test := range badData {
		q := make(url.Values)
		q.Set("token", test)

		e := newTestEcho(t)
		req := httptest.NewRequest(http.MethodGet, "/boot/ipxe?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

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
	host.Kickstart = true
	err = h.DB.StoreHost(host)
	assert.NoError(err)

	token, err := model.NewBootToken(host.ID.String(), host.Interfaces[0].MAC.String())
	assert.NoError(err)

	q := make(url.Values)
	q.Set("token", token)

	e := newTestEcho(t)
	req := httptest.NewRequest(http.MethodGet, "/boot/ipxe?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(TokenRequired(h.Ipxe)(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.True(strings.HasPrefix(rec.Body.String(), "#!ipxe"))
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
	host.Kickstart = true
	err = h.DB.StoreHost(host)
	assert.NoError(err)

	hostBad := tests.HostFactory.MustCreate().(*model.Host)

	token, err := model.NewBootToken(hostBad.ID.String(), hostBad.Interfaces[0].MAC.String())
	assert.NoError(err)

	q := make(url.Values)
	q.Set("token", token)

	e := newTestEcho(t)
	req := httptest.NewRequest(http.MethodGet, "/boot/ipxe?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

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
