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
	fragment := app.Group("/fragments")
	api := app.Group("/api")

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
			"SearchList":  hostList,
			"CurrentPath": c.Path(),
		})
		if sess.Get("role") == "disabled" {
			c.Response().Header.Add("HX-Trigger", `{"toast-error": "Your account is disabled. Please ask an Administrator to activate your account."}`)
		}
		if err != nil {
			log.Error(err)
		}
		return c.Next()
	})

	app.Static("/", "frontend/public")

	app.Get("/", h.Index)

	app.Get("/login", h.Login)
	api.Post("/auth/login", h.LoginUser)
	api.Post("/auth/logout", h.LogoutUser)

	app.Get("/register", h.Register)
	api.Post("/auth/register", h.RegisterUser)

	app.Get("/host/:host", auth, h.Host)
	api.Post("/host", auth, h.EditHost)
	api.Delete("/host", auth, h.DeleteHost)
	api.Post("/host/add", auth, h.HostAdd)
	api.Post("/host/add2", auth, h.HostAdd2)
	api.Post("/switch/mac", auth, h.SwitchMac)
	fragment.Get("/hostAddModal", auth, h.HostAddModal)
	fragment.Put("/hostAddModalList", auth, h.HostAddModalList)
	fragment.Put("/hostAddModalInterfaces", auth, h.HostAddModalInterfaces)

	app.Get("/floorplan", auth, h.Floorplan)
	fragment.Get("/floorplan/table", auth, h.floorplanTable)
	fragment.Get("/floorplan/addHost", auth, h.floorplanAddHost)
	fragment.Get("/floorplan/interfaces", auth, h.floorplanInterfaces)

	app.Get("/rack/:rack", auth, h.Rack)

	app.Get("/users", auth, h.Users)
	api.Post("/users", auth, h.UsersPost)
	fragment.Get("/userTable", auth, h.userTable)

	api.Get("/search", auth, h.Search)

	api.Post("/bmc/reboot", auth, h.RebootHost)
	api.Post("/bmc/configure", auth, h.BmcConfigure)
}
