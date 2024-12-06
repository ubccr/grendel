// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/buntstore"
)

func newTestDB(t *testing.T) store.Store {
	db, err := buntstore.New(":memory:")
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
