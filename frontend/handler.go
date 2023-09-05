package frontend

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ubccr/grendel/model"
)

type Handler struct {
	DB model.DataStore
}

func NewHandler(db model.DataStore) (*Handler, error) {
	h := &Handler{
		DB: db,
	}

	return h, nil
}

func (h *Handler) SetupRoutes(app *fiber.App) {

	keyAuth := AuthMiddleware()

	app.Static("/", "frontend/public")

	app.Get("/", h.Index)
	app.Get("/login", h.Login)
	app.Get("/register", h.Register)
	app.Get("/host/:host", keyAuth, h.Host)
	app.Get("/floorplan", keyAuth, h.Floorplan)
	app.Get("/rack/:rack", keyAuth, h.Rack)

	fragment := app.Group("/fragment")
	fragment.Get("/hostAddModal", keyAuth, h.HostAddModal)
	fragment.Put("/hostAddModalList", keyAuth, h.HostAddModalList)

	api := app.Group("/api")
	api.Post("/auth/login", h.LoginUser)
	api.Post("/auth/logout", h.LogoutUser)
	api.Post("/auth/register", h.RegisterUser)
	api.Post("/host", keyAuth, h.EditHost)
	api.Post("/host/add", keyAuth, h.HostAdd)
	api.Post("/bmc/reboot", keyAuth, h.RebootHost)
	api.Post("/bmc/configure", keyAuth, h.BmcConfigure)
	api.Post("/switch/mac", keyAuth, h.SwitchMac)
}
