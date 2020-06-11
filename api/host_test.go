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

package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/model"
)

func TestHostAdd(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	e := newEcho()

	addHostJSON := "[" + string(tests.TestHostJSON) + "]"

	req := httptest.NewRequest(http.MethodPost, "/host", strings.NewReader(addHostJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.HostAdd(c)) {
		assert.Equal(http.StatusCreated, rec.Code)
		assert.True(gjson.Get(rec.Body.String(), "hosts").Exists())
		assert.Equal(int64(1), gjson.Get(rec.Body.String(), "hosts").Int())
	}
}

func TestHostAddWrongContentType(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	e := newEcho()

	addHostJSON := "this is not json"

	req := httptest.NewRequest(http.MethodPost, "/host", strings.NewReader(addHostJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HostAdd(c)
	if assert.Error(err) {
		he, ok := err.(*echo.HTTPError)
		if ok {
			assert.Equal(http.StatusBadRequest, he.Code)
		}
		e.HTTPErrorHandler(err, c)
		assert.Equal(http.StatusBadRequest, rec.Code)
		assert.True(gjson.Get(rec.Body.String(), "message").Exists())
	}
}

func TestHostAddInvalid(t *testing.T) {
	assert := assert.New(t)

	addHostJSON := "[" + string(tests.TestHostJSON) + "]"

	badData := make([]string, 4)
	// Test invalid IP Address
	badData[0], _ = sjson.Set(addHostJSON, "0.interfaces.0.ip", "bad ip")
	// Test missing required Name
	badData[1], _ = sjson.Set(addHostJSON, "0.name", "")
	// Test invalid json payload
	badData[2] = "{bad json]"
	// Test single host (needs to be a list)
	badData[3] = string(tests.TestHostJSON)

	h := &Handler{newTestDB(t)}

	e := newEcho()

	for _, test := range badData {
		req := httptest.NewRequest(http.MethodPost, "/host", strings.NewReader(test))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.HostAdd(c)
		if assert.Error(err) {
			he, ok := err.(*echo.HTTPError)
			if ok {
				assert.Equal(http.StatusBadRequest, he.Code)
			}
			e.HTTPErrorHandler(err, c)
			assert.Equal(http.StatusBadRequest, rec.Code)
			assert.True(gjson.Get(rec.Body.String(), "message").Exists())
		}
	}
}

func TestHostList(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	size := 10
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		err := h.DB.StoreHost(host)
		assert.NoError(err)
	}

	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/host/list", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.HostList(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(size, len(gjson.Parse(rec.Body.String()).Array()))
	}
}

func TestHostFind(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := h.DB.StoreHost(host)
		assert.NoError(err)
	}

	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/host/find/tux-[05-14]", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.HostFind(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(10, len(gjson.Parse(rec.Body.String()).Array()))
	}
}

func TestHostFindNone(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := h.DB.StoreHost(host)
		assert.NoError(err)
	}

	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/host/find/cpn-[05-14]", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.HostFind(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(0, len(gjson.Parse(rec.Body.String()).Array()))
	}
}

func TestHostFindInvalidNodeSet(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/host/find/tux-[05[-14]", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HostFind(c)
	if assert.Error(err) {
		he, ok := err.(*echo.HTTPError)
		if ok {
			assert.Equal(http.StatusBadRequest, he.Code)
		}
		e.HTTPErrorHandler(err, c)
		assert.Equal(http.StatusBadRequest, rec.Code)
		assert.True(gjson.Get(rec.Body.String(), "message").Exists())
	}
}
