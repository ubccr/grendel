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

func TestHostFindByTags(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	size := 10
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Name = fmt.Sprintf("tux-%02d", i)
		if (i % 2) == 0 {
			host.Tags = []string{"d13", "ib"}
		} else if (i % 2) != 0 {
			host.Tags = []string{"d16", "noib"}
		}
		err := h.DB.StoreHost(host)
		assert.NoError(err)
	}

	testResults := map[string]int{
		"/host/tags/ib":           5,
		"/host/tags/ib,noib":      10,
		"/host/tags/doesnotexist": 0,
	}

	e := newEcho()

	for path, count := range testResults {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(h.HostFindByTags(c)) {
			assert.Equal(http.StatusOK, rec.Code)
			assert.Equal(count, len(gjson.Parse(rec.Body.String()).Array()))
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/host/tags/", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HostFindByTags(c)
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

func TestHostProvision(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Provision = false
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := h.DB.StoreHost(host)
		assert.NoError(err)
	}

	hostList, err := h.DB.Hosts()
	if assert.NoError(err) {
		count := 0
		for _, host := range hostList {
			if host.Provision {
				count++
			}
		}
		assert.Equal(0, count)
	}

	e := newEcho()

	req := httptest.NewRequest(http.MethodPut, "/host/provision/tux-[05-14]", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.HostProvision(c)) {
		assert.Equal(http.StatusOK, rec.Code)
	}

	hostList, err = h.DB.Hosts()
	if assert.NoError(err) {
		count := 0
		for _, host := range hostList {
			if host.Provision {
				count++
			}
		}
		assert.Equal(10, count)
	}
}

func TestHostTag(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Provision = false
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := h.DB.StoreHost(host)
		assert.NoError(err)
	}

	hostList, err := h.DB.Hosts()
	if assert.NoError(err) {
		count := 0
		for _, host := range hostList {
			if len(host.Tags) > 0 {
				count++
			}
		}
		assert.Equal(0, count)
	}

	e := newEcho()

	req := httptest.NewRequest(http.MethodPut, "/host/tag/tux-[05-14]", nil)
	params := req.URL.Query()
	params.Add("tags", "ib")
	req.URL.RawQuery = params.Encode()

	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.HostTag(c)) {
		assert.Equal(http.StatusOK, rec.Code)
	}

	ns, err := h.DB.FindTags([]string{"ib"})
	if assert.NoError(err) {
		assert.Equal(10, ns.Len())
	}
}

func TestHostDelete(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	size := 20
	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		host.Provision = false
		host.Name = fmt.Sprintf("tux-%02d", i)
		err := h.DB.StoreHost(host)
		assert.NoError(err)
	}

	e := newEcho()

	req := httptest.NewRequest(http.MethodDelete, "/host/find/tux-[05-09]", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.HostDelete(c)) {
		assert.Equal(http.StatusOK, rec.Code)
	}

	hostList, err := h.DB.Hosts()
	if assert.NoError(err) {
		assert.Equal(15, len(hostList))
	}
}
