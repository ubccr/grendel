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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/model"
)

var (
	statusJSON = `{"status":"up"}`
)

func newTestDB(t *testing.T) model.DataStore {
	db, err := model.NewDataStore(":memory:")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	return db
}

func TestCreateUser(t *testing.T) {
	h := &Handler{newTestDB(t)}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, h.Index(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, statusJSON, strings.TrimSpace(rec.Body.String()))
	}
}
