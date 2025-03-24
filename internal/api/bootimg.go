// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"fmt"
	"slices"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/pkg/model"
)

type BootImageAddRequest struct {
	BootImages model.BootImageList `json:"boot_images"`
}

func (h *Handler) BootImageAdd(c fuego.ContextWithBody[BootImageAddRequest]) (*GenericResponse, error) {
	images, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to parse body",
		}
	}

	err = h.DB.StoreBootImages(images.BootImages)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to add image(s)",
		}
	}

	var names []string
	for _, image := range images.BootImages {
		names = append(names, image.Name)
	}

	h.writeEvent(c.Context(), "Success", fmt.Sprintf("Successfully saved image(s): %s", strings.Join(names, ",")))

	return &GenericResponse{
		Title:   "Success",
		Detail:  "successfully added image(s)",
		Changed: len(images.BootImages),
	}, nil
}

func (h *Handler) BootImageList(c fuego.ContextNoBody) (model.BootImageList, error) {
	imageList, err := h.DB.BootImages()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to get images",
		}
	}

	return imageList, nil
}

func (h *Handler) BootImageFind(c fuego.ContextNoBody) (model.BootImageList, error) {
	// TODO: this should be handled in the DB
	names := strings.Split(c.QueryParam("names"), ",")

	images, err := h.DB.BootImages()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to find images",
		}
	}

	var imageList model.BootImageList
	for _, image := range images {
		if slices.Contains(names, image.Name) {
			imageList = append(imageList, image)
		}
	}

	return imageList, nil
}

func (h *Handler) BootImageDelete(c fuego.ContextNoBody) (*GenericResponse, error) {
	names := strings.Split(c.QueryParam("name"), ",")

	err := h.DB.DeleteBootImages(names)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to delete images",
		}
	}

	h.writeEvent(c.Context(), "Success", fmt.Sprintf("Successfully deleted image(s): %s", names))

	return &GenericResponse{
		Title:   "Success",
		Detail:  "successfully deleted image(s)",
		Changed: len(names),
	}, err
}
