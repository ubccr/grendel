// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"errors"
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/pkg/model"
)

func (h *Handler) Restore(c fuego.ContextWithBody[model.DataDump]) (*GenericResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to parse restore body",
		}
	}

	err = h.DB.RestoreFrom(body)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to restore db",
		}
	}

	log.Infof("Database restored successfully")
	h.writeEvent(c.Context(), "Success", "Successfully restored DB.")

	return &GenericResponse{
		Title:  "Success",
		Detail: fmt.Sprintf("restored db: hosts=%d images=%d users=%d roles=%d", len(body.Hosts), len(body.Images), len(body.Users), len(body.Roles)),
	}, nil
}

func (h *Handler) Dump(c fuego.ContextNoBody) (*model.DataDump, error) {
	nodeList, err := h.DB.Hosts()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to dump nodes from db",
		}
	}
	imageList, err := h.DB.BootImages()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to dump boot images from db",
		}
	}
	userList, err := h.DB.GetUsers()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to dump users from db",
		}
	}

	roleList, err := h.DB.GetRoles()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to dump roles from db",
		}
	}
	userCreatedRoles := make(model.RoleViewList, 0)

	for _, role := range roleList {
		_, err := model.RoleFromString(role.Name)
		if errors.Is(err, model.ErrInvalidRole) {
			userCreatedRoles = append(userCreatedRoles, role)
		}
	}

	dump := &model.DataDump{
		Hosts:  nodeList,
		Images: imageList,
		Users:  userList,
		Roles:  userCreatedRoles,
	}

	return dump, nil
}
