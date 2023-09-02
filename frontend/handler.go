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

	app.Static("/favicon.ico", "frontend/public/favicon.ico")
	app.Static("/backgrounds/large-triangles-ub.svg", "frontend/public/backgrounds/large-triangles-ub.svg")
	app.Static("/tailwind.css", "frontend/public/tailwind.css")

	app.Get("/", h.Index)
	app.Get("/login", h.Login)
	app.Get("/register", h.Register)
	app.Get("/host/:host", keyAuth, h.Host)
	app.Get("/floorplan", keyAuth, h.Floorplan)
	app.Get("/rack/:rack", keyAuth, h.Rack)
	app.Get("/grendel/add", keyAuth, h.GrendelAdd)

	api := app.Group("/api")
	api.Post("/auth/login", h.LoginUser)
	api.Post("/auth/logout", h.LogoutUser)
	api.Post("/auth/register", h.RegisterUser)
	api.Post("/host", keyAuth, h.EditHost)
	api.Post("/bmc/reboot", keyAuth, h.RebootHost)
	api.Post("/bmc/configure", keyAuth, h.BmcConfigure)

}
