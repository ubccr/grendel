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
	"errors"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

func (h *Handler) HostAdd(c echo.Context) error {
	var hosts model.HostList

	if !strings.HasPrefix(c.Request().Header.Get(echo.HeaderContentType), echo.MIMEApplicationJSON) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid content type")
	}

	if err := c.Bind(&hosts); err != nil {
		return err
	}

	log.Infof("Attempting to add %d hosts", len(hosts))

	for _, host := range hosts {
		err := c.Validate(host)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid data").SetInternal(err)
		}
	}

	err := h.DB.StoreHosts(hosts)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save hosts").SetInternal(err)
	}

	log.Infof("Added %d hosts successfully", len(hosts))

	res := map[string]interface{}{
		"hosts": len(hosts),
	}
	return c.JSON(http.StatusCreated, res)
}

func (h *Handler) HostList(c echo.Context) error {
	hostList, err := h.DB.Hosts()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch hosts").SetInternal(err)
	}
	return c.JSON(http.StatusOK, hostList)
}

func (h *Handler) HostFind(c echo.Context) error {
	_, nodesetString := path.Split(c.Request().URL.Path)

	nodeset, err := nodeset.NewNodeSet(nodesetString)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid nodeset").SetInternal(err)
	}

	log.Infof("Got nodeset: %s", nodeset.String())

	hostList, err := h.DB.FindHosts(nodeset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find hosts").SetInternal(err)
	}
	return c.JSON(http.StatusOK, hostList)
}

func (h *Handler) HostFindByTags(c echo.Context) error {
	_, tagStr := path.Split(c.Request().URL.Path)
	if tagStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid list of tags")
	}

	tags := strings.Split(tagStr, ",")
	if len(tags) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid list of tags")
	}

	log.Infof("Got tags: %s", tags)

	ns, err := h.DB.FindTags(tags)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return c.JSON(http.StatusOK, model.HostList{})
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find nodeset with tags").SetInternal(err)
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch hosts by tag").SetInternal(err)
	}
	return c.JSON(http.StatusOK, hostList)
}

func (h *Handler) hostSetProvision(c echo.Context, provision bool) error {
	_, nodesetString := path.Split(c.Request().URL.Path)

	nodeset, err := nodeset.NewNodeSet(nodesetString)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid nodeset").SetInternal(err)
	}

	log.Infof("Got nodeset: %s", nodeset.String())

	err = h.DB.ProvisionHosts(nodeset, provision)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "No hosts found in nodeset").SetInternal(err)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update hosts provision property").SetInternal(err)
	}

	log.Infof("Set %d hosts provision=%s", nodeset.Len(), strconv.FormatBool(provision))

	res := map[string]interface{}{
		"hosts": nodeset.Len(),
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) HostProvision(c echo.Context) error {
	return h.hostSetProvision(c, true)
}

func (h *Handler) HostUnprovision(c echo.Context) error {
	return h.hostSetProvision(c, false)
}

func (h *Handler) hostSetTags(c echo.Context, remove bool) error {
	_, nodesetString := path.Split(c.Request().URL.Path)

	nodeset, err := nodeset.NewNodeSet(nodesetString)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid nodeset").SetInternal(err)
	}

	tagStr := c.QueryParam("tags")
	if tagStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid list of tags")
	}

	tags := strings.Split(tagStr, ",")
	if len(tags) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid list of tags")
	}

	log.Infof("Got tags: %s", tags)
	log.Infof("Got nodeset: %s", nodeset.String())

	if remove {
		err = h.DB.UntagHosts(nodeset, tags)
	} else {
		err = h.DB.TagHosts(nodeset, tags)
	}

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "No hosts found in nodeset").SetInternal(err)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update hosts tag property").SetInternal(err)
	}

	verb := "add"
	if remove {
		verb = "remove"
	}

	log.Infof("Set %d hosts %s tags=%s", nodeset.Len(), verb, tags)

	res := map[string]interface{}{
		"hosts":  nodeset.Len(),
		"tags":   tags,
		"action": verb,
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) HostTag(c echo.Context) error {
	return h.hostSetTags(c, false)
}

func (h *Handler) HostUntag(c echo.Context) error {
	return h.hostSetTags(c, true)
}
