package frontend

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

func (h *Handler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.gohtml", nil)
}

func (h *Handler) Register(c echo.Context) error {
	return c.Render(http.StatusOK, "register.gohtml", nil)
}

func (h *Handler) Login(c echo.Context) error {
	return c.Render(http.StatusOK, "login.gohtml", nil)
}

type HostPageData struct {
	Host       *model.Host
	BootImages model.BootImageList
	Firmware   []string
}

func (h *Handler) Host(c echo.Context) error {
	reqHost, _ := nodeset.NewNodeSet(c.Param("host"))

	host, _ := h.DB.FindHosts(reqHost)
	bootImages, _ := h.DB.BootImages()

	fw := make([]string, 0)
	for _, i := range firmware.BuildToStringMap {
		fw = append(fw, i)
	}

	data := HostPageData{
		Host:       host[0],
		BootImages: bootImages,
		Firmware:   fw,
	}
	return c.Render(http.StatusOK, "host.gohtml", data)
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
	for i := 28; i >= 5; i-- {
		cols = append(cols, fmt.Sprintf("%02d", i))
	}

	data := map[string]interface{}{
		"Rows":  rows,
		"Cols":  cols,
		"Racks": racks,
	}
	return c.Render(http.StatusOK, "floorplan.gohtml", data)
}

func (h *Handler) Rack(c echo.Context) error {
	n, _ := h.DB.FindTags([]string{c.Param("rack")})
	hosts, _ := h.DB.FindHosts(n)

	u := make([]string, 0)
	// TODO: move min and max rack u to grendel.toml
	for i := 42; i >= 3; i-- {
		u = append(u, fmt.Sprintf("%02d", i))
	}
	data := map[string]interface{}{
		"u":     u,
		"Hosts": hosts,
		"Rack":  c.Param("rack"),
	}

	return c.Render(http.StatusOK, "rack.gohtml", data)
}

func (h *Handler) GrendelAdd(c echo.Context) error {
	

	return c.Render(http.StatusOK, "grendelAdd.gohtml", nil)
}