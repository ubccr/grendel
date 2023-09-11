package frontend

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/ubccr/grendel/model"
)

type Handler struct {
	DB    model.DataStore
	Store *session.Store
}

func NewHandler(db model.DataStore, Store *session.Store) (*Handler, error) {
	h := &Handler{
		DB:    db,
		Store: Store,
	}

	return h, nil
}

func (h *Handler) SetupRoutes(app *fiber.App) {

	auth := h.EnforceAuthMiddleware()
	app.Use(func(c *fiber.Ctx) error {
		hostList, err := h.DB.Hosts()
		if err != nil {
			log.Error(err)
		}

		sess, _ := h.Store.Get(c)
		err = c.Bind(fiber.Map{
			"Auth": fiber.Map{
				"Authenticated": sess.Get("authenticated"),
				"User":          sess.Get("user"),
				"Role":          sess.Get("role"),
			},
			"SearchList": hostList,
			"CurrentPath": c.Path(),
		})
		if err != nil {
			log.Error(err)
		}
		return c.Next()
	})

	app.Static("/", "frontend/public")

	app.Get("/", h.Index)
	app.Get("/login", h.Login)
	app.Get("/register", h.Register)
	app.Get("/host/:host", auth, h.Host)
	app.Get("/floorplan", auth, h.Floorplan)
	app.Get("/rack/:rack", auth, h.Rack)

	fragment := app.Group("/fragments")
	fragment.Get("/hostAddModal", auth, h.HostAddModal)
	fragment.Put("/hostAddModalList", auth, h.HostAddModalList)
	fragment.Put("/hostAddModalInterfaces", auth, h.HostAddModalInterfaces)

	api := app.Group("/api")
	api.Post("/auth/login", h.LoginUser)
	api.Post("/auth/logout", h.LogoutUser)
	api.Post("/auth/register", h.RegisterUser)
	api.Post("/host", auth, h.EditHost)
	api.Delete("/host", auth, h.DeleteHost)
	api.Post("/host/add", auth, h.HostAdd)
	api.Post("/bmc/reboot", auth, h.RebootHost)
	api.Post("/bmc/configure", auth, h.BmcConfigure)
	api.Post("/switch/mac", auth, h.SwitchMac)
	api.Get("/search", auth, h.Search)
}
