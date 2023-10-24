package frontend

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) EnforceAuthMiddleware() func(f *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		sess, _ := h.Store.Get(c)
		if sess.Get("authenticated") == nil {
			msg := "Authentication required. Please login to continue."
			c.Response().Header.Add("HX-Trigger", fmt.Sprintf(`{"toast-error": "%s"}`, msg))
			return c.Status(fiber.StatusUnauthorized).Redirect("/login")
		}
		return c.Next()
	}
}
