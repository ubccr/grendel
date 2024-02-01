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
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

type promServiceDisovery struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func (h *Handler) ServiceDiscovery(c echo.Context) error {
	tag := c.Param("tag")
	if tag == "" {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	port := 0
	err := echo.PathParamsBinder(c).Int("port", &port).BindError()
	if err != nil || port == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid port").SetInternal(err)
	}

	labels := make(map[string]string, len(c.QueryParams()))
	for key, _ := range c.QueryParams() {
		labels[key] = c.QueryParam(key)
	}

	sd := make([]*promServiceDisovery, 0)
	nodeExporter := &promServiceDisovery{
		Targets: make([]string, 0),
		Labels:  labels,
	}

	hosts, err := h.DB.Hosts()
	if err != nil {
		log.Error("failed to fetch all hosts for service discovery")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find hosts").SetInternal(err)
	}

	for _, h := range hosts {
		if !h.HasTags(tag) {
			continue
		}

		bootNic := h.BootInterface()
		if bootNic != nil {
			nodeExporter.Targets = append(nodeExporter.Targets, fmt.Sprintf("%s:%d", bootNic.HostName(), port))
		}
	}

	sd = append(sd, nodeExporter)
	c.Response().Header().Set("X-Prometheus-Refresh-Interval-Seconds", viper.GetString("provision.prometheus_sd_refresh_interval"))
	return c.JSON(http.StatusOK, sd)
}
