// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// skip auth if bound to unix socket
		if viper.IsSet("api.socket_path") {
			next.ServeHTTP(w, r)
			return
		}

		var rawToken string
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			rawToken = authHeader
		}
		authCookie, _ := r.Cookie("Authorization")
		if authCookie != nil {
			rawToken = authCookie.Value
		}

		if rawToken == "" {
			err := fuego.UnauthorizedError{
				Err:    fmt.Errorf("authentication error ip=%s", r.RemoteAddr),
				Title:  "Error",
				Detail: "failed to authenticate",
			}
			ErrorSerializer(w, r, err)
			log.Error(err.Unwrap().Error())
			return
		}

		token := strings.TrimPrefix(rawToken, "Bearer ")

		claims, err := ParseToken(token, viper.GetString("api.secret"))
		if err != nil {
			err := fuego.HTTPError{
				Status: http.StatusBadRequest,
				Err:    err,
				Title:  "Error",
				Detail: "Failed to verify token",
			}
			ErrorSerializer(w, r, err)
			log.Error(err.Unwrap().Error())
			return
		}

		user, err := h.DB.GetUserByName(claims.username)
		if err != nil {
			err := fuego.HTTPError{
				Status: http.StatusBadRequest,
				Err:    err,
				Title:  "Error",
				Detail: "Invalid Username",
			}
			ErrorSerializer(w, r, err)
			log.Error(err.Unwrap().Error())
			return
		}

		if !user.Enabled {
			err := fuego.HTTPError{
				Status: http.StatusBadRequest,
				Err:    fmt.Errorf("user %s is not enabled", user.Username),
				Title:  "Error",
				Detail: "Account disabled, please ask an admin to enable it",
			}
			ErrorSerializer(w, r, err)
			log.Error(err.Unwrap().Error())
			return
		}

		validRoles, err := h.DB.GetRolesByRoute(r.Method, r.URL.Path)
		if err != nil {
			err := fuego.HTTPError{
				Status: http.StatusInternalServerError,
				Err:    err,
				Title:  "Error",
				Detail: "Failed to retrieve permissions",
			}
			ErrorSerializer(w, r, err)
			log.Error(err.Unwrap().Error())
			return
		}

		if !slices.Contains(*validRoles, claims.role) || !slices.Contains(*validRoles, user.Role) {
			err := fuego.HTTPError{
				Status: http.StatusForbidden,
				Err:    fmt.Errorf("account does not have the required permissions to access this endpoint: user=%s, method=%s, path=%s, role=%s, validRoles=%s", claims.username, r.Method, r.URL.Path, claims.role, strings.Join(*validRoles, ",")),
				Title:  "Error",
				Detail: "Assigned role does not have the required permissions to access this endpoint",
			}
			ErrorSerializer(w, r, err)
			log.Error(err.Unwrap().Error())
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyUsername, claims.username)
		ctx = context.WithValue(ctx, ContextKeyRole, claims.role)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("api request: method=%s route=%s ip=%s", r.Method, r.URL, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(enabled bool) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enabled {
				cors.New(cors.Options{
					AllowedOrigins:   []string{"*"},
					AllowedMethods:   []string{"GET", "PATCH", "POST", "PUT", "DELETE", "OPTIONS"},
					AllowedHeaders:   []string{"*"},
					AllowCredentials: true,
				}).HandlerFunc(w, r)
			}
			next.ServeHTTP(w, r)
		})
	}
}
