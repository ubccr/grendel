// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

type promServiceDiscovery struct {
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

	sd := make([]*promServiceDiscovery, 0)
	nodeExporter := &promServiceDiscovery{
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
