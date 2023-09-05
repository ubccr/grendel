package frontend

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/ubccr/grendel/firmware"
)

func (h *Handler) HostAddModal(f *fiber.Ctx) error {
	bootImages, _ := h.DB.BootImages()

	fw := make([]string, 0)
	for _, i := range firmware.BuildToStringMap {
		fw = append(fw, i)
	}

	return f.Render("fragments/hostModal", fiber.Map{
		"Rack":       f.FormValue("rack"),
		"HostUs":     f.FormValue("hosts"),
		"Firmwares":  fw,
		"BootImages": bootImages,
	}, "")
}

func (h *Handler) HostAddModalList(f *fiber.Ctx) error {
	rack := f.FormValue("Rack")
	prefix := f.FormValue("Prefix")
	uArr := strings.Split(f.FormValue("Us"), ",")

	type hostStruct struct {
		Host     string
		MgmtPort string
		CorePort string
		MgmtMac  string
		CoreMac  string
	}
	hostArr := make([]hostStruct, len(uArr))
	hostList := make([]string, len(uArr))

	for i, v := range uArr {
		host := fmt.Sprintf("%s-%s-%s", prefix, rack, v)
		hostList[i] = host
		hostArr[i] = hostStruct{
			Host:     host,
			MgmtPort: f.FormValue(fmt.Sprintf("%s:Mgmt", host), ""),
			CorePort: f.FormValue(fmt.Sprintf("%s:Core", host), ""),
			MgmtMac:  f.FormValue(fmt.Sprintf("%s:MgmtMac", host), ""),
			CoreMac:  f.FormValue(fmt.Sprintf("%s:CoreMac", host), ""),
		}
	}

	return f.Render("fragments/hostAddModalList", fiber.Map{
		"Hosts":    hostArr,
		"HostList": strings.Join(hostList, ","),
	}, "")
}
