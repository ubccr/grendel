package frontend

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

func AuthMiddleware() func(f *fiber.Ctx) error {
	return keyauth.New(keyauth.Config{
		KeyLookup: "cookie:Authorization",
		Validator: validateKey,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err == keyauth.ErrMissingOrMalformedAPIKey {
				msg := "Authentication required: missing or malformed API key"
				c.Response().Header.Add("HX-Trigger", fmt.Sprintf(`{"toast-error": "%s"}`, msg))
				return c.Status(fiber.StatusUnauthorized).SendString(msg)
			}
			c.Response().Header.Add("HX-Trigger", `{"toast-error": "Invalid or expired API Key"}`)
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid or expired API Key")
		},
	})
}

func validateKey(c *fiber.Ctx, key string) (bool, error) {
	_, r, err := Verify(key)

	if err != nil {
		return false, err
	} else if r == "disabled" {
		return false, nil
	}

	return true, nil
}
