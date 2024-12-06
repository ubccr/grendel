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

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/ubccr/grendel/internal/tests"
	"github.com/ubccr/grendel/pkg/model"
)

func TestRestore(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	e := newEcho()

	size := 10
	adminUsername := "admin"
	adminPassword := "pass1234123"

	dump := model.DataDump{
		Users:  make([]model.User, 0),
		Hosts:  make(model.HostList, 0),
		Images: make(model.BootImageList, 0),
	}

	role, err := h.DB.StoreUser(adminUsername, adminPassword)
	assert.NoError(err)
	assert.Equal("admin", role)
	role, err = h.DB.StoreUser("user", "test1234")
	assert.NoError(err)
	assert.Equal("disabled", role)

	for i := 0; i < size; i++ {
		host := tests.HostFactory.MustCreate().(*model.Host)
		dump.Hosts = append(dump.Hosts, host)
	}

	beforeHostList, err := h.DB.Hosts()
	assert.NoError(err)
	assert.Equal(0, len(beforeHostList))

	for i := 0; i < size; i++ {
		image := tests.BootImageFactory.MustCreate().(*model.BootImage)
		dump.Images = append(dump.Images, image)
	}

	users, err := h.DB.GetUsers()
	assert.NoError(err)
	for _, u := range users {
		dump.Users = append(dump.Users, u)
	}

	dataDumpBytes, err := json.Marshal(&dump)
	assert.NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/restore", strings.NewReader(string(dataDumpBytes)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.Restore(c)) {
		assert.Equal(http.StatusCreated, rec.Code)
		assert.True(gjson.Get(rec.Body.String(), "ok").Exists())
	}

	hostList, err := h.DB.Hosts()
	assert.NoError(err)
	assert.Equal(len(dump.Hosts), len(hostList))

	imageList, err := h.DB.BootImages()
	assert.NoError(err)
	assert.Equal(len(dump.Images), len(imageList))

	users2, err := h.DB.GetUsers()
	assert.NoError(err)
	assert.Equal(len(dump.Users), len(users2))

	authenticated, role, err := h.DB.VerifyUser(adminUsername, adminPassword)
	assert.NoError(err)
	assert.Equal("admin", role)
	assert.Equal(true, authenticated)
}
