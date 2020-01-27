package api

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
	e.GET("/", h.Index).Name = "index"

	v1 := e.Group("/v1/")
	v1.POST("host/add", h.HostAdd)
	v1.GET("host/list", h.HostList)
	v1.GET("host/find/*", h.HostFind)
}

func (h *Handler) Index(c echo.Context) error {
	resp := map[string]interface{}{
		"status": "up",
	}
	return c.JSON(http.StatusOK, resp)
}
