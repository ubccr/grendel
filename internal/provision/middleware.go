// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ubccr/grendel/pkg/model"
)

func TokenRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Param("token")
		if token == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "missing token")
		}

		claims, err := model.ParseBootToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid token").SetInternal(err)
		}

		c.Set(ContextKeyToken, claims)

		return next(c)
	}
}
