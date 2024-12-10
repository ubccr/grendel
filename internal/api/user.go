// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) UserList(c echo.Context) error {
	users, err := h.DB.GetUsers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch users").SetInternal(err)
	}
	return c.JSON(http.StatusOK, users)
}
