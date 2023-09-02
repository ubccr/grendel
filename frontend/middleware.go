package frontend

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

func AuthMiddleware() func(f *fiber.Ctx) error {
	return keyauth.New(keyauth.Config{
		KeyLookup: "cookie:Authorization",
		Validator: validateKey,
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
