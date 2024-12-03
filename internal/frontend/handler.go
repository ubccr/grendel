package frontend

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/bmc"
	"github.com/ubccr/grendel/pkg/model"
)

//go:embed public
var embedFS embed.FS

type EventStruct struct {
	Time        string
	User        string
	Severity    string
	Message     string
	JobMessages []bmc.JobMessage
}

type Handler struct {
	DB     model.DataStore
	Store  *session.Store
	Events []EventStruct
}

func NewHandler(db model.DataStore, Store *session.Store) (*Handler, error) {
	h := &Handler{
		DB:     db,
		Store:  Store,
		Events: []EventStruct{},
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
			"Events":      h.Events,
		})
		if sess.Get("role") == "disabled" {
			c.Response().Header.Add("HX-Trigger", `{"toast-error": "Your account is disabled. Please ask an Administrator to activate your account."}`)
		}
		if err != nil {
			log.Error(err)
		}
		return c.Next()
	})

	public, err := fs.Sub(embedFS, "public")
	if err != nil {
		log.Error("failed to load public files")
	}

	app.Use("/static", filesystem.New(filesystem.Config{
		Root:   http.FS(public),
		Browse: false,
	}))

	if viper.IsSet("frontend.favicon") {
		app.Use(favicon.New(favicon.Config{
			File: viper.GetString("frontend.favicon"),
		}))
	} else {
		app.Use(favicon.New(favicon.Config{
			File:       "favicon.ico",
			FileSystem: http.FS(public),
		}))
	}

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
	api.Post("/host/import", auth, h.importHost)
	api.Post("/host/inventory", auth, h.AddInventoryHost)

	fragment.Get("/interfaces", auth, h.interfaces)
	fragment.Get("/bonds", auth, h.bonds)

	app.Get("/floorplan", auth, h.Floorplan)
	fragment.Get("/floorplan/table", auth, h.floorplanTable)
	fragment.Get("/floorplan/modal", auth, h.floorplanModal)

	app.Get("/rack/:rack", auth, h.Rack)
	fragment.Get("/rack/:rack/table", auth, h.rackTable)
	fragment.Get("/rack/:rack/add/modal", auth, h.rackAddModal)
	fragment.Post("/rack/:rack/add/table", auth, h.rackAddTable)

	app.Get("/nodes", auth, h.nodes)
	fragment.Get("/nodes", auth, h.nodesTable)

	app.Get("/power", auth, h.power)
	fragment.Get("/power/panels", auth, h.powerPanels)

	fragment.Put("/actions", auth, h.actions)
	api.Post("/bulkHostAdd", auth, h.bulkHostAdd)

	api.Patch("/hosts/provision", auth, h.provisionHosts)
	api.Patch("/hosts/tags", auth, h.tagHosts)
	api.Patch("/hosts/image", auth, h.imageHosts)
	api.Get("/hosts/export/:hosts", auth, h.exportHosts)
	api.Get("/hosts/inventory/:hosts", auth, h.exportInventory)

	app.Get("/users", admin, h.Users)
	api.Post("/users", admin, h.usersPost)
	fragment.Get("/users/table", admin, h.usersTable)
	api.Delete("/user/:username", admin, h.deleteUser)

	api.Get("/search", auth, h.Search)

	fragment.Get("/events", auth, h.events)

	app.Get("/status", h.status)

	api.Post("/bmc/powerCycle", auth, h.bmcPowerCycle)
	api.Post("/bmc/powerCycleBmc", auth, h.bmcPowerCycleBmc)
	api.Post("/bmc/clearSel", auth, h.bmcClearSel)
	api.Post("/bmc/configure/auto", auth, h.bmcConfigureAuto)
	api.Post("/bmc/configure/import", auth, h.bmcConfigureImport)
}
