// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
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
	Username string `json:"username" description:"username shown in logs, does not need to be a valid user in the DB" example:"user1:CLI"`
	Role     string `json:"role" description:"type of model.Role, valid options: disabled, user, admin" example:"admin"`
	Expire   string `json:"expire" description:"string parsed by time.ParseDuration, examples include: infinite, 8h, 30m, 20s" example:"infinite"`
}

type AuthTokenReponse struct {
	Token string `json:"token"`
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

	if body.Expire != "infinite" {
		exp, err := time.ParseDuration(body.Expire)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Title:  "Authentication Error",
				Detail: "failed to parse expire time",
			}
		}
		claims[TokenExpire] = exp
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
