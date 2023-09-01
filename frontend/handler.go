package frontend

import (
	"net/http"

	"github.com/labstack/echo/v4"
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

func (h *Handler) SetupRoutes(e *echo.Echo) {

	e.File("/favicon.ico", "frontend/public/favicon.ico")
	e.File("/backgrounds/large-triangles-ub.svg", "frontend/public/backgrounds/large-triangles-ub.svg")
	e.File("/tailwind.css", "frontend/public/tailwind.css")

	e.GET("/", h.Index)
	e.GET("/login", h.Login)
	e.GET("/register", h.Register)
	e.GET("/host/:host", h.Host, AuthMiddleware())
	e.GET("/floorplan", h.Floorplan, AuthMiddleware())
	e.GET("/rack/:rack", h.Rack, AuthMiddleware())
	e.GET("/grendel/add", h.GrendelAdd, AuthMiddleware())

	api := e.Group("/api")
	api.POST("/auth/login", h.LoginUser)
	api.POST("/auth/logout", h.LogoutUser)
	api.POST("/auth/register", h.RegisterUser)
	api.POST("/host", h.EditHost, AuthMiddleware())
	// api.PATCH("/provision/:nodeset", h.Provision, AuthMiddleware())
	api.POST("/bmc/reboot", h.RebootHost, AuthMiddleware())
	api.POST("/bmc/configure", h.BmcConfigure, AuthMiddleware())

	e.HTTPErrorHandler = customHTTPErrorHandler
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	log.Error(err)
	message := err.Error()

	if message == "code=400, message=missing key in cookies" {
		message = "Authentication failed! Please login."
	}

	if err := c.Render(code, "error.gohtml", message); err != nil {
		log.Error(err)
	}
}
