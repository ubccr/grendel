// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/pkg/model"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthSignupRequest struct {
	Username string `json:"username" validate:"required,min=2"`
	Password string `json:"password" validate:"required,min=8"`
}

type AuthResponse struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	Token    string `json:"token"`
	Expire   int64  `json:"expire"`
}

type AuthTokenRequest struct {
	Username string `json:"username" description:"username"`
	Role     string `json:"role" description:"type of model.Role, valid options: disabled, user, admin" example:"admin"`
	Expire   string `json:"expire" description:"string parsed by time.ParseDuration, examples include: infinite, 8h, 30m, 20s" example:"infinite"`
}

type AuthTokenReponse struct {
	Token string `json:"token"`
}

type AuthResetRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

var (
	expireDuration = time.Duration(8) * time.Hour
)

func (h *Handler) AuthSignin(c fuego.ContextWithBody[AuthRequest]) (*AuthResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Authentication Error",
			Detail: "failed to parse auth body",
		}
	}

	authenticated, role, err := h.DB.VerifyUser(body.Username, body.Password)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Authentication Error",
			Detail: "failed to login, invalid credentials",
		}
	}
	if !authenticated {
		return nil, fuego.HTTPError{
			Err:    errors.New("invalid credentials"),
			Title:  "Authentication Error",
			Detail: "failed to login, invalid credentials",
		}
	}

	exp := time.Now().Add(expireDuration)

	claims := jwt.MapClaims{
		TokenUsername: body.Username,
		TokenRole:     role,
		TokenExpire:   exp.Unix(),
	}

	token, err := NewToken(claims, viper.GetString("api.secret"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Authentication Error",
			Detail: "failed to create token",
		}
	}

	c.SetCookie(http.Cookie{
		Name:     "Authorization",
		Value:    "Bearer " + token,
		Expires:  exp,
		Secure:   viper.IsSet("api.cert"),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	return &AuthResponse{
		Username: body.Username,
		Role:     role,
		Expire:   exp.UnixMilli(),
	}, nil
}

func (h *Handler) AuthSignup(c fuego.ContextWithBody[AuthSignupRequest]) (*AuthResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Authentication Error",
			Detail: "failed to parse auth body",
		}
	}

	role, err := h.DB.StoreUser(body.Username, body.Password)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Authentication Error",
			Detail: "failed to store user: " + body.Username,
		}
	}

	exp := time.Now().Add(expireDuration)

	claims := jwt.MapClaims{
		TokenUsername: body.Username,
		TokenRole:     role,
		TokenExpire:   exp.Unix(),
	}

	token, err := NewToken(claims, viper.GetString("api.secret"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Authentication Error",
			Detail: "failed to create token",
		}
	}

	c.SetCookie(http.Cookie{
		Name:     "Authorization",
		Value:    "Bearer " + token,
		Expires:  exp,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	return &AuthResponse{
		Username: body.Username,
		Role:     role,
		Expire:   exp.UnixMilli(),
	}, nil
}

func (h *Handler) AuthSignout(c fuego.ContextNoBody) (*GenericResponse, error) {
	c.SetCookie(http.Cookie{
		Name:     "Authorization",
		Value:    "",
		Expires:  time.Now(),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	return &GenericResponse{
		Title:   "Success",
		Detail:  "successfully signed out",
		Changed: 1,
	}, nil
}

func (h *Handler) AuthToken(c fuego.ContextWithBody[AuthTokenRequest]) (*AuthTokenReponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Authentication Error",
			Detail: "failed to parse auth body",
		}
	}

	claims := jwt.MapClaims{
		TokenUsername: body.Username,
		TokenRole:     body.Role,
	}

	// skip access control if running on a unix socket
	if !viper.IsSet("api.socket_path") {
		tokenRole, err := h.DB.GetRolesByName(body.Role)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Error",
				Detail: "failed to get token role",
			}
		}
		contextRole, ok := c.Context().Value(ContextKeyRole).(string)
		if !ok {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Error",
				Detail: "failed to parse role context",
			}
		}
		contextUsername, ok := c.Context().Value(ContextKeyUsername).(string)
		if !ok {
			return nil, fuego.HTTPError{
				Err:    errors.New("failed to parse username from context"),
				Title:  "Error",
				Detail: "failed to parse username context",
			}
		}

		requestUser, err := h.DB.GetUserByName(body.Username)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Error",
				Detail: "failed to get token username",
			}
		}
		requestedRole, err := h.DB.GetRolesByName(requestUser.Role)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Error",
				Detail: "failed to get token role",
			}
		}

		if contextRole != model.RoleAdmin.String() && contextUsername != body.Username {
			return nil, fuego.HTTPError{
				Err:    errors.New("token username does not match request username"),
				Title:  "Error",
				Detail: "Failed to create token because of mismatched username. Only admins are able to create tokens with different usernames.",
			}
		}

		var requiredPermissions []string
		for _, r := range requestedRole.UnassignedPermissionList {
			match := false
			for _, t := range tokenRole.PermissionList {
				if t.Method == r.Method && t.Path == r.Path {
					match = true
				}
			}
			if match {
				requiredPermissions = append(requiredPermissions, fmt.Sprintf("%s:%s", r.Method, r.Path))
			}
		}

		if len(requiredPermissions) > 0 {
			return nil, fuego.HTTPError{
				Err:    fmt.Errorf("failed to create token with role %s because of missing permissions: %s", body.Role, strings.Join(requiredPermissions, ",")),
				Title:  "Error",
				Detail: "Failed to create token because of missing permissions, assign a role with lesser or equal permissions to the users role",
			}
		}
	}

	if body.Expire != "infinite" {
		exp, err := time.ParseDuration(body.Expire)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Authentication Error",
				Detail: "failed to parse expire time",
			}
		}
		claims[TokenExpire] = time.Now().Add(exp).Unix()
	}

	token, err := NewToken(claims, viper.GetString("api.secret"))
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Authentication Error",
			Detail: "failed to create token",
		}
	}

	return &AuthTokenReponse{
		Token: token,
	}, nil
}

func (h *Handler) AuthReset(c fuego.ContextWithBody[AuthResetRequest]) (*GenericResponse, error) {
	body, err := c.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Authentication Error",
			Detail: "failed to parse body",
		}
	}

	username, ok := c.Context().Value(ContextKeyUsername).(string)
	if !ok {
		return nil, errors.New("failed to parse username from context")
	}

	authenticated, _, err := h.DB.VerifyUser(username, body.CurrentPassword)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Authentication Error",
			Detail: "failed to login, invalid credentials",
		}
	}

	if !authenticated {
		return nil, fuego.HTTPError{
			Err:    errors.New("invalid credentials"),
			Title:  "Authentication Error",
			Detail: "failed to login, invalid credentials",
		}
	}

	_, err = h.DB.StoreUser(username, body.NewPassword)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Title:  "Error",
			Detail: "Failed to update credentials",
		}
	}

	return &GenericResponse{
		Title:   "Success",
		Detail:  "succesfully updated credentials",
		Changed: 1,
	}, nil
}
