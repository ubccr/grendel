// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package frontend

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) EnforceAuthMiddleware() func(f *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		sess, _ := h.Store.Get(c)
		if sess.Get("authenticated") == nil || sess.Get("role") == "disabled" {
			msg := "Authentication required. Please login to continue."
			c.Response().Header.Add("HX-Trigger", fmt.Sprintf(`{"toast-error": "%s"}`, msg))
			return c.Status(fiber.StatusUnauthorized).Redirect("/login")
		}
		return c.Next()
	}
}
func (h *Handler) EnforceAdminMiddleware() func(f *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		sess, _ := h.Store.Get(c)
		if sess.Get("authenticated") == nil || sess.Get("role") != "admin" {
			msg := "Admin role required."
			c.Response().Header.Add("HX-Trigger", fmt.Sprintf(`{"toast-error": "%s"}`, msg))
			return c.SendStatus(fiber.StatusForbidden)
		}
		return c.Next()
	}
}
