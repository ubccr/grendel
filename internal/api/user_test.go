// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestUserList(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	role, err := h.DB.StoreUser("admin", "pass1234")
	assert.NoError(err)
	assert.Equal("admin", role)

	e := newEcho()

	req := httptest.NewRequest(http.MethodGet, "/user/list", nil)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.UserList(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(1, len(gjson.Parse(rec.Body.String()).Array()))
	}
}
