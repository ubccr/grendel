// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/pkg/model"
)

func (h *Handler) BootImageAdd(c echo.Context) error {
	var images model.BootImageList

	if !strings.HasPrefix(c.Request().Header.Get(echo.HeaderContentType), echo.MIMEApplicationJSON) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid content type")
	}

	if err := c.Bind(&images); err != nil {
		return err
	}

	log.Infof("Attempting to add %d boot images", len(images))

	for _, image := range images {
		err := c.Validate(image)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid data").SetInternal(err)
		}
	}

	err := h.DB.StoreBootImages(images)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save boot images").SetInternal(err)
	}

	log.Infof("Stored %d images successfully", len(images))

	res := map[string]interface{}{
		"images": len(images),
	}
	return c.JSON(http.StatusCreated, res)
}

func (h *Handler) BootImageList(c echo.Context) error {
	imageList, err := h.DB.BootImages()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch images").SetInternal(err)
	}
	return c.JSON(http.StatusOK, imageList)
}

func (h *Handler) BootImageFind(c echo.Context) error {
	name := c.Param("name")
	imageList := make(model.BootImageList, 0)

	image, err := h.DB.LoadBootImage(name)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "image not found").SetInternal(err)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch image").SetInternal(err)
	}

	imageList = append(imageList, image)
	return c.JSON(http.StatusOK, imageList)
}

func (h *Handler) BootImageDelete(c echo.Context) error {
	name := c.Param("name")

	// TODO add support for deleting more than one image
	err := h.DB.DeleteBootImages([]string{name})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete image").SetInternal(err)
	}

	res := map[string]interface{}{
		"images": 1,
	}

	return c.JSON(http.StatusOK, res)
}
