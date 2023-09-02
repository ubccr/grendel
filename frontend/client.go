package frontend

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

func (h *Handler) Index(f *fiber.Ctx) error {
	return f.Render("index", nil)
}

func (h *Handler) Register(f *fiber.Ctx) error {
	return f.Render("register", nil)
}

func (h *Handler) Login(f *fiber.Ctx) error {
	return f.Render("login", nil)
}

type HostPageData struct {
	Host       *model.Host
	BootImages model.BootImageList
	Firmware   []string
}

func (h *Handler) Host(f *fiber.Ctx) error {
	reqHost, err := nodeset.NewNodeSet(f.Params("host"))
	if err != nil {
		return ToastError(f, fmt.Errorf("invalid host"), "Invalid host")
	}

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
	return f.Render("host", data)
}
func (h *Handler) Floorplan(f *fiber.Ctx) error {
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
	return f.Render("floorplan", data)
}

func (h *Handler) Rack(f *fiber.Ctx) error {
	n, err := h.DB.FindTags([]string{f.Params("rack")})
	if err != nil {
		return ToastError(f, err, "Failed to find hosts tagged with rack")
	}

	hosts, err := h.DB.FindHosts(n)
	if err != nil {
		return ToastError(f, err, "Failed to find hosts")
	}

	u := make([]string, 0)
	// TODO: move min and max rack u to grendel.toml
	for i := 42; i >= 3; i-- {
		u = append(u, fmt.Sprintf("%02d", i))
	}
	data := map[string]interface{}{
		"u":     u,
		"Hosts": hosts,
		"Rack":  f.FormValue("rack"),
	}

	return f.Render("rack", data)
}

func (h *Handler) GrendelAdd(f *fiber.Ctx) error {

	return f.Render("grendelAdd", nil)
}
