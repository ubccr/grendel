// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ubccr/grendel/internal/store"
)

type Handler struct {
	DB store.Store
}

func NewHandler(db store.Store) (*Handler, error) {
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
	v1.DELETE("host/find/*", h.HostDelete)
	v1.GET("host/tags/*", h.HostFindByTags)
	v1.PUT("host/tag/*", h.HostTag)
	v1.PUT("host/untag/*", h.HostUntag)
	v1.PUT("host/provision/*", h.HostProvision)
	v1.PUT("host/unprovision/*", h.HostUnprovision)

	v1.POST("bootimage", h.BootImageAdd)
	v1.GET("bootimage/find/:name", h.BootImageFind)
	v1.DELETE("bootimage/find/:name", h.BootImageDelete)
	v1.GET("bootimage/list", h.BootImageList)
	v1.GET("user/list", h.UserList)
	v1.POST("restore", h.Restore)
}

func (h *Handler) Index(c echo.Context) error {
	resp := map[string]interface{}{
		"status": "up",
	}
	return c.JSON(http.StatusOK, resp)
}
