package frontend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/segmentio/ksuid"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
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
	e.File("/tailwind.config.js", "frontend/public/tailwind.config.js")

	e.GET("/", h.Index)
	e.GET("/host/:host", h.Host)
	e.GET("/floorplan", h.Floorplan)
	e.GET("/rack/:rack", h.Rack)

	e.PATCH("/provision/:nodeset", h.Provision)
	e.POST("/host", h.EditHost)

}

func (h *Handler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}
func (h *Handler) Host(c echo.Context) error {
	reqHost, _ := nodeset.NewNodeSet(c.Param("host"))

	host, _ := h.DB.FindHosts(reqHost)
	bootImages, _ := h.DB.BootImages()

	fw := make([]string, 0)
	for _, i := range firmware.BuildToStringMap {
		fw = append(fw, i)
	}

	data := map[string]interface{}{
		"Host":       host[0],
		"BootImages": bootImages,
		"Firmware":   fw,
	}
	return c.Render(http.StatusOK, "host.html", data)
}
func (h *Handler) Floorplan(c echo.Context) error {
	hosts, _ := h.DB.Hosts()
	racks := map[string]int{}
	for _, host := range hosts {
		rack := strings.Split(host.Name, "-")[1]
		racks[rack] += 1
	}

	// TODO: make this configurable in grendel.toml
	rows := make([]string, 0)
	for i := 'f'; i <= 'v'; i++ {
		rows = append(rows, fmt.Sprintf("%c", i))
	}
	cols := make([]string, 0)
	for i := 28; i >= 3; i-- {
		cols = append(cols, fmt.Sprintf("%02d", i))
	}

	data := map[string]interface{}{
		"Rows":  rows,
		"Cols":  cols,
		"Racks": racks,
	}
	return c.Render(http.StatusOK, "floorplan.html", data)
}
func (h *Handler) Rack(c echo.Context) error {
	n, _ := h.DB.FindTags([]string{c.Param("rack")})
	hosts, _ := h.DB.FindHosts(n)

	u := make([]string, 0)
	// TODO: move min and max rack u to grendel.toml
	for i := 42; i > 4; i-- {
		u = append(u, fmt.Sprintf("%02d", i))
	}
	data := map[string]interface{}{
		"u":     u,
		"Hosts": hosts,
		"Rack":  c.Param("rack"),
	}
	return c.Render(http.StatusOK, "rack.html", data)
}

func (h *Handler) Provision(c echo.Context) error {
	reqHost, _ := nodeset.NewNodeSet(c.Param("nodeset"))
	host, _ := h.DB.FindHosts(reqHost)
	h.DB.ProvisionHosts(reqHost, !host[0].Provision)

	return c.HTML(http.StatusOK, fmt.Sprintf("%t", !host[0].Provision))
}

type FormData struct {
	ID         string
	Name       string
	Provision  string
	Firmware   string
	BootImage  string
	Tags       string
	Interfaces string
}

func (h *Handler) EditHost(c echo.Context) error {
	formHost := new(FormData)
	if err := c.Bind(formHost); err != nil {
		log.Warn(err)
		return c.HTML(http.StatusBadRequest, "Failed to bind type to request body")
	}
	if err := c.Validate(formHost); err != nil {
		log.Warn(err)
		return c.HTML(http.StatusBadRequest, "Failed to bind type to request body")
	}

	id, _ := ksuid.Parse(formHost.ID)

	provision := false
	if formHost.Provision == "on" {
		provision = true
	}
	var ifaces []*model.NetInterface
	json.Unmarshal([]byte(formHost.Interfaces), &ifaces)

	newHost := model.Host{
		ID:         id,
		Name:       formHost.Name,
		Provision:  provision,
		Firmware:   firmware.NewFromString(formHost.Firmware),
		BootImage:  formHost.BootImage,
		Tags:       strings.Split(formHost.Tags, ","),
		Interfaces: ifaces,
	}

	h.DB.StoreHost(&newHost)
	return c.HTML(http.StatusOK, "<h1 class='text-green-500'>Successfully updated host!</h1>")
}
