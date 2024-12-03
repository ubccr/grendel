// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

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
	"github.com/ubccr/grendel/pkg/model"
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

	req := httptest.NewRequest(http.MethodGet, "/bootimage/find/centos14", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/bootimage/find/:name")
	c.SetParamNames("name")
	c.SetParamValues("centos14")

	if assert.NoError(h.BootImageFind(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(1, len(gjson.Parse(rec.Body.String()).Array()))
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

	req := httptest.NewRequest(http.MethodGet, "/bootimage/find/centos50", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/bootimage/find/:name")
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

func TestBootImageDelete(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodDelete, "/bootimage/find/centos14", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/bootimage/find/:name")
	c.SetParamNames("name")
	c.SetParamValues("centos14")

	if assert.NoError(h.BootImageDelete(c)) {
		assert.Equal(http.StatusOK, rec.Code)
	}

	bootImages, err := h.DB.BootImages()
	if assert.NoError(err) {
		assert.Equal(19, len(bootImages))
	}
}
