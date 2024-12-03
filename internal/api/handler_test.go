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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/ubccr/grendel/pkg/model"
)

func newTestDB(t *testing.T) model.DataStore {
	db, err := model.NewDataStore(":memory:")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	return db
}

func TestStatus(t *testing.T) {
	assert := assert.New(t)

	h := &Handler{newTestDB(t)}

	e := newEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(h.Index(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal("up", gjson.Get(rec.Body.String(), "status").String())
	}
}
