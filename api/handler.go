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

	"github.com/labstack/echo/v4"
	"github.com/ubccr/grendel/model"
)

type Handler struct {
	DB model.DataStore
}

func NewHandler(db model.DataStore) (*Handler, error) {
	h := &Handler{
		DB: db,
	}

	return h, nil
}

func (h *Handler) SetupRoutes(e *echo.Echo) {
	e.GET("/", h.Index).Name = "index"

	v1 := e.Group("/v1/")
	v1.POST("host", h.HostAdd)
	v1.GET("host/list", h.HostList)
	v1.GET("host/find/*", h.HostFind)

	v1.POST("bootimage", h.BootImageAdd)
	v1.GET("bootimage/find/:name", h.BootImageFind)
	v1.GET("bootimage/list", h.BootImageList)
}

func (h *Handler) Index(c echo.Context) error {
	resp := map[string]interface{}{
		"status": "up",
	}
	return c.JSON(http.StatusOK, resp)
}
