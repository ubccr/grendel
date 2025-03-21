package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/pkg/model"
)

func authMiddleware(next http.Handler) http.Handler {
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
			fuego.SendError(w, r, err)
			log.Error(err.Unwrap().Error())
			return
		}

		token := strings.TrimPrefix(rawToken, "Bearer ")

		claims, err := ParseToken(token, viper.GetString("api.secret"))
		if err != nil {
			err := fuego.HTTPError{
				Err:    fmt.Errorf("authentication error ip=%s err=%s", r.RemoteAddr, err),
				Title:  "Error",
				Detail: "failed to verify token",
			}
			fuego.SendError(w, r, err)
			log.Error(err.Unwrap().Error())
			return
		}

		if claims.role == model.RoleDisabled.String() {
			err := fuego.HTTPError{
				Err:    fmt.Errorf("authentication error, account is disabled ip=%s, user=%s", r.RemoteAddr, claims.username),
				Title:  "Error",
				Detail: "account is disabled",
			}
			fuego.SendError(w, r, err)
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
