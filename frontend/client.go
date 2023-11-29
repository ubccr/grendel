package frontend

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) Index(f *fiber.Ctx) error {
	return f.Render("index", fiber.Map{
		"Title": "Grendel",
	})
}

func (h *Handler) Register(f *fiber.Ctx) error {
	return f.Render("register", fiber.Map{
		"Title": "Grendel - Register",
	})
}

func (h *Handler) Login(f *fiber.Ctx) error {
	return f.Render("login", fiber.Map{
		"Title": "Grendel - Login",
	})
}

func (h *Handler) Floorplan(f *fiber.Ctx) error {
	return f.Render("floorplan", fiber.Map{
		"Title": "Grendel - Floorplan",
	})
}

func (h *Handler) Rack(f *fiber.Ctx) error {
	rack := f.Params("rack")
	return f.Render("rack", fiber.Map{
		"Title": fmt.Sprintf("Grendel - %s", rack),
		"Rack":  rack,
	})
}

func (h *Handler) Host(f *fiber.Ctx) error {
	host := f.Params("host")
	return f.Render("host", fiber.Map{
		"Title":    fmt.Sprintf("Grendel - %s", host),
		"HostName": host,
	})
}

func (h *Handler) Users(f *fiber.Ctx) error {
	return f.Render("users", fiber.Map{})
}
