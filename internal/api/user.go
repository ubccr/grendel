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

type UserStoreRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserStoreResponse struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type UserRoleRequest struct {
	Role string `json:"role" description:"type of model.Role, valid options: disabled, user, admin" example:"admin"`
}

type UserEnableRequest struct {
	Enabled bool `json:"enabled"`
}

func (h *Handler) UserList(c fuego.ContextNoBody) ([]model.User, error) {
	users, err := h.DB.GetUsers()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to get users",
		}
	}

	if c.QueryParam("usernames") != "" {
		usersFilter := strings.Split(c.QueryParam("usernames"), ",")
		filteredUsers := make([]model.User, 0)
		for _, user := range users {
			if slices.Contains(usersFilter, user.Username) {
				filteredUsers = append(filteredUsers, user)
			}
		}
		return filteredUsers, nil
	}

	return users, nil
}

func (h *Handler) UserDelete(c fuego.ContextNoBody) (*GenericResponse, error) {
	users := strings.Split(c.PathParam("usernames"), ",")

	for _, user := range users {
		err := h.DB.DeleteUser(user)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Error",
				Detail: "failed to delete user: " + user,
			}
		}
	}

	h.writeEvent(c.Context(), "Success", fmt.Sprintf("Successfully deleted user(s): %s", strings.Join(users, ", ")))

	return &GenericResponse{
		Title:   "Success",
		Detail:  "successfully deleted user(s)",
		Changed: len(users),
	}, nil
}

func (h *Handler) UserRole(c fuego.ContextWithBody[UserRoleRequest]) (*GenericResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to parse user role body",
		}
	}

	users := strings.Split(c.PathParam("usernames"), ",")

	for _, user := range users {
		err := h.DB.UpdateUserRole(user, body.Role)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Error",
				Detail: "failed to update user: " + user,
			}
		}
	}

	h.writeEvent(c.Context(), "Success", fmt.Sprintf("Successfully updated user(s) role: %s", strings.Join(users, ", ")))

	return &GenericResponse{
		Title:   "Success",
		Detail:  "successfully edited user(s) role",
		Changed: len(users),
	}, nil
}

func (h *Handler) UserEnable(c fuego.ContextWithBody[UserEnableRequest]) (*GenericResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to parse user enable body",
		}
	}

	users := strings.Split(c.PathParam("usernames"), ",")

	for _, user := range users {
		err := h.DB.UpdateUserEnabled(user, body.Enabled)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Error",
				Detail: "failed to update user: " + user,
			}
		}
	}

	h.writeEvent(c.Context(), "Success", fmt.Sprintf("Successfully updated user(s) enable flag: users=%s enabled=%t", strings.Join(users, ", "), body.Enabled))

	return &GenericResponse{
		Title:   "Success",
		Detail:  "successfully edited user(s) enabled flag",
		Changed: len(users),
	}, nil
}

func (h *Handler) UserStore(c fuego.ContextWithBody[UserStoreRequest]) (*UserStoreResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to parse user verify body",
		}
	}

	role, err := h.DB.StoreUser(body.Username, body.Password)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "failed to store user: " + body.Username,
		}
	}

	return &UserStoreResponse{
		Username: body.Username,
		Role:     role,
	}, nil
}
