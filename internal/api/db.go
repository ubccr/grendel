// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"errors"
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/settings"
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

	if err := h.restoreSettings(body.Settings, actor(c.Context())); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to restore settings",
		}
	}

	log.Infof("Database restored successfully")
	h.audit(c.Context(), model.SeveritySuccess, "db.restore", "Restored the database", "")

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

	rows, _, err := h.DB.Settings()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to dump settings from db",
		}
	}
	stored := []model.StoredSetting{}
	for _, row := range rows {
		// a generated secret is this install's own, and restoring it elsewhere would make two servers accept each other's tokens
		if !settings.Dumped(row.Key) {
			continue
		}
		stored = append(stored, model.StoredSetting{Key: row.Key, Value: string(row.Value)})
	}

	dump := &model.DataDump{
		Hosts:    nodeList,
		Images:   imageList,
		Users:    userList,
		Roles:    userCreatedRoles,
		Settings: stored,
	}

	return dump, nil
}
