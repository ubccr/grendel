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
	admin := h.EnforceAdminMiddleware()
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
	fragment.Get("/host/:host/form", auth, h.hostForm)
	api.Post("/host", auth, h.EditHost)
	api.Delete("/host", auth, h.DeleteHost)

	fragment.Get("/interfaces", auth, h.interfaces)

	app.Get("/floorplan", auth, h.Floorplan)
	fragment.Get("/floorplan/table", auth, h.floorplanTable)
	fragment.Get("/floorplan/modal", auth, h.floorplanModal)

	app.Get("/rack/:rack", auth, h.Rack)
	fragment.Get("/rack/:rack/table", auth, h.rackTable)
	fragment.Get("/rack/:rack/actions", auth, h.rackActions)
	fragment.Get("/rack/:rack/add/modal", auth, h.rackAddModal)
	fragment.Post("/rack/:rack/add/table", auth, h.rackAddTable)

	api.Post("/bulkHostAdd", auth, h.bulkHostAdd)

	api.Patch("/hosts/provision", auth, h.provisionHosts)
	api.Patch("/hosts/tags", auth, h.tagHosts)
	api.Patch("/hosts/image", auth, h.imageHosts)
	api.Get("/hosts/export/:hosts", auth, h.exportHosts)

	app.Get("/users", admin, h.Users)
	api.Post("/users", admin, h.usersPost)
	fragment.Get("/users/table", admin, h.usersTable)
	api.Delete("/user/:username", admin, h.deleteUser)

	api.Get("/search", auth, h.Search)

	api.Post("/bmc/reboot", auth, h.RebootHost)
	api.Post("/bmc/configure/auto", auth, h.bmcConfigureAuto)
	api.Post("/bmc/configure/import", auth, h.bmcConfigureImport)
}
