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
	return f.Render("index", fiber.Map{
		"Title": "Grendel",
	})
}

func (h *Handler) Register(f *fiber.Ctx) error {
	return f.Render("register", fiber.Map{
		"Title": "Grendel - Register",
	})
}

func (h *Handler) Login(f *fiber.Ctx) error {
	return f.Render("login", fiber.Map{
		"Title": "Grendel - Login",
	})
}

func (h *Handler) Host(f *fiber.Ctx) error {
	reqHost, err := nodeset.NewNodeSet(f.Params("host"))
	if err != nil {
		return ToastError(f, fmt.Errorf("invalid host"), "Invalid host")
	}

	host, err := h.DB.FindHosts(reqHost)
	if err != nil || len(host) == 0 {
		return ToastError(f, err, "Failed to find host")
	}
	bootImages, err := h.DB.BootImages()
	if err != nil {
		return ToastError(f, err, "Failed to load boot images")
	}

	fw := make([]string, 0)
	for _, i := range firmware.BuildToStringMap {
		fw = append(fw, i)
	}

	return f.Render("host", fiber.Map{
		"Title":      fmt.Sprintf("Grendel - %s", host[0].Name),
		"Host":       host[0],
		"BootImages": bootImages,
		"Firmware":   fw,
	})
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

	return f.Render("floorplan", fiber.Map{
		"Title": "Grendel - Floorplan",
		"Rows":  rows,
		"Cols":  cols,
		"Racks": racks,
	})
}

func (h *Handler) Rack(f *fiber.Ctx) error {
	rack := f.Params("rack")

	n, err := h.DB.FindTags([]string{rack})
	if err != nil {
		return ToastError(f, err, "Failed to find hosts tagged with rack")
	}

	hosts, err := h.DB.FindHosts(n)
	if err != nil {
		return ToastError(f, err, "Failed to find hosts")
	}

	var filtered model.HostList
	for _, v := range hosts {
		if v.HostType() == "power" && !v.HasAnyTags("1u", "2u") {
			continue
		}

		filtered = append(filtered, v)
	}

	u := make([]string, 0)
	// TODO: move min and max rack u to grendel.toml
	for i := 42; i >= 3; i-- {
		u = append(u, fmt.Sprintf("%02d", i))
	}

	return f.Render("rack", fiber.Map{
		"Title": fmt.Sprintf("Grendel - %s", rack),
		"u":     u,
		"Hosts": filtered,
		"Rack":  rack,
	})
}

func (h *Handler) Users(f *fiber.Ctx) error {
	users, err := h.DB.GetUsers()
	if err != nil {
		return ToastError(f, err, "Failed to load users")
	}

	return f.Render("users", fiber.Map{
		"Users": users,
	})
}
