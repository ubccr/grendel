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

	for _, h := range hosts {
		if !h.HasTags(tag) {
			continue
		}

		bootNic := h.BootInterface()
		if bootNic == nil {
			continue
		}
		labels := make(map[string]string, 0)

		for _, tag := range h.Tags {
			if strings.Contains(tag, "panel") {
				labels["panel"] = strings.Replace(tag, "panel:", "", 1)
			} else if tag == "faculty" || tag == "ubhpc" {
				labels["cluster"] = tag
			} else if strings.Contains(tag, "partition") {
				labels["partition"] = strings.Replace(tag, "partition:", "", 1)
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
