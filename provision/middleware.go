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

package provision

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ubccr/grendel/model"
)

func TokenRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.QueryParam("token")
		if token == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "missing token")
		}

		claims, err := model.ParseBootToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid token")
		}

		c.Set(ContextKeyToken, claims)

		return next(c)
	}
}
