// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"fmt"
	"maps"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func (h *Handler) PDUServiceDiscovery(c echo.Context) error {
	tag := c.Param("tag")
	if tag == "" {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	port := 0
	err := echo.PathParamsBinder(c).Int("port", &port).BindError()
	if err != nil || port == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid port").SetInternal(err)
	}

	sd := make([]*promServiceDiscovery, 0)

	hosts, err := h.DB.Hosts()
	if err != nil {
		log.Error("failed to fetch all hosts for service discovery")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find hosts").SetInternal(err)
	}

	allLabels := make(map[string]string, 0)
	for key := range c.QueryParams() {
		allLabels[key] = c.QueryParam(key)
	}

	for _, h := range hosts {
		if !h.HasTags(tag) {
			continue
		}

		bootNic := h.BootInterface()
		if bootNic == nil {
			continue
		}
		labels := make(map[string]string, 0)
		maps.Copy(labels, allLabels)

		for _, tag := range h.Tags {
			if strings.Contains(tag, "panel") {
				labels["panel"] = strings.Replace(tag, "panel:", "", 1)
			} else if strings.Contains(tag, "rack_type") {
				labels["rack_type"] = strings.Replace(tag, "rack_type:", "", 1)
			} else if strings.Contains(tag, "generation") {
				labels["generation"] = strings.Replace(tag, "generation:", "", 1)
			}
		}
		nameSlice := strings.Split(h.Name, "-")
		if len(nameSlice) > 1 {
			labels["rack"] = nameSlice[1]
		}

		nodeExporter := &promServiceDiscovery{
			Targets: []string{fmt.Sprintf("%s:%d", bootNic.HostName(), port)},
			Labels:  labels,
		}
		sd = append(sd, nodeExporter)
	}

	c.Response().Header().Set("X-Prometheus-Refresh-Interval-Seconds", viper.GetString("provision.prometheus_sd_refresh_interval"))
	return c.JSON(http.StatusOK, sd)
}
