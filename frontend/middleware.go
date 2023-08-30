package frontend

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"strings"
)

func AuthMiddleware() echo.MiddlewareFunc {
	return middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup: "cookie:Authorization",
		Validator: func(key string, c echo.Context) (bool, error) {
			keySplice := strings.Split(key, " ")
			token := keySplice[1]

			_, r, err := Verify(token)

			if err != nil {
				return false, err
			} else if r == "disabled" {
				return false, nil
			}

			return true, nil
		},
	})
}
