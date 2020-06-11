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

func TestBootImageAdd(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	e := newEcho()

	addBootImageJSON := "[" + string(tests.TestBootImageJSON) + "]"

	req := httptest.NewRequest(http.MethodPost, "/bootimage", strings.NewReader(addBootImageJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.BootImageAdd(c)) {
		assert.Equal(http.StatusCreated, rec.Code)
		assert.True(gjson.Get(rec.Body.String(), "images").Exists())
		assert.Equal(int64(1), gjson.Get(rec.Body.String(), "images").Int())
	}
}

func TestBootImageAddWrongContentType(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	e := newEcho()

	addBootImageJSON := "this is not json"

	req := httptest.NewRequest(http.MethodPost, "/bootimage", strings.NewReader(addBootImageJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.BootImageAdd(c)
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

func TestBootImageAddInvalid(t *testing.T) {
	assert := assert.New(t)

	addBootImageJSON := "[" + string(tests.TestBootImageJSON) + "]"

	badData := make([]string, 4)
	// Test missing required Kernel
	badData[0], _ = sjson.Set(addBootImageJSON, "0.kernel", "")
	// Test missing required Name
	badData[1], _ = sjson.Set(addBootImageJSON, "0.name", "")
	// Test invalid json payload
	badData[2] = "{bad json]"
	// Test single boot image (needs to be a list)
	badData[3] = string(tests.TestBootImageJSON)

	h := &Handler{newTestDB(t)}

	e := newEcho()

	for _, test := range badData {
		req := httptest.NewRequest(http.MethodPost, "/bootimage", strings.NewReader(test))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.BootImageAdd(c)
		if assert.Errorf(err, "No error for bad data: %s", test) {
			he, ok := err.(*echo.HTTPError)
			if ok {
				assert.Equalf(http.StatusBadRequest, he.Code, "wrong http error status code for bad data: %s", test)
			}
			e.HTTPErrorHandler(err, c)
			assert.Equalf(http.StatusBadRequest, rec.Code, "wrong status code for bad data: %s", test)
			assert.Truef(gjson.Get(rec.Body.String(), "message").Exists(), "missing error message for bad data: %s", test)
		}
	}
}

func TestBootImageList(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	size := 10
	for i := 0; i < size; i++ {
		image := tests.BootImageFactory.MustCreate().(*model.BootImage)
		err := h.DB.StoreBootImage(image)
		assert.NoError(err)
	}

	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/bootimage/list", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.BootImageList(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(size, len(gjson.Parse(rec.Body.String()).Array()))
	}
}

func TestBootImageFind(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	size := 20
	for i := 0; i < size; i++ {
		image := tests.BootImageFactory.MustCreate().(*model.BootImage)
		image.Name = fmt.Sprintf("centos%02d", i)
		err := h.DB.StoreBootImage(image)
		assert.NoError(err)
	}

	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/bootimage/centos14", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/bootimage/:name")
	c.SetParamNames("name")
	c.SetParamValues("centos14")

	if assert.NoError(h.BootImageFind(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal("centos14", gjson.Get(rec.Body.String(), "name").String())
	}
}

func TestBootImageFindNone(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	size := 20
	for i := 0; i < size; i++ {
		image := tests.BootImageFactory.MustCreate().(*model.BootImage)
		image.Name = fmt.Sprintf("centos%02d", i)
		err := h.DB.StoreBootImage(image)
		assert.NoError(err)
	}

	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/bootimage/centos50", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/bootimage/:name")
	c.SetParamNames("name")
	c.SetParamValues("centos50")

	err := h.BootImageFind(c)
	if assert.Error(err) {
		he, ok := err.(*echo.HTTPError)
		if ok {
			assert.Equal(http.StatusNotFound, he.Code)
		}
		e.HTTPErrorHandler(err, c)
		assert.Equal(http.StatusNotFound, rec.Code)
	}
}
