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
		"Defaults": fiber.Map{
			"Core": fiber.Map{
				"Ifname": "eno12399",
				"Domain": "core.ccr.buffalo.edu",
				"Mtu":    9000,
			},
			"Mgmt": fiber.Map{
				"Ifname": "",
				"Domain": "mgmt.ccr.buffalo.edu",
				"Mtu":    1500,
			},
		},
	}, "")
}

func (h *Handler) HostAddModalList(f *fiber.Ctx) error {
	rack := f.FormValue("Rack")
	prefix := f.FormValue("Prefix")
	Us := f.FormValue("Us")
	var uArr []string

	if Us != "" {
		uArr = strings.Split(Us, ",")
	}

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
			MgmtPort: f.FormValue(fmt.Sprintf("%s:Mgmt", host), v),
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

func (h *Handler) HostAddModalInterfaces(f *fiber.Ctx) error {
	mgmtSubnet := f.FormValue("Mgmt:Subnet")
	coreSubnet := f.FormValue("Core:Subnet")
	hostList := f.FormValue("HostList")
	hostArr := strings.Split(hostList, ",")

	mgmtRange := make([]string, len(hostArr))
	coreRange := make([]string, len(hostArr))
	err := error(nil)

	if mgmtSubnet != "" {
		mgmtRange, err = h.newHostIPs(mgmtSubnet)
		if err != nil {
			return ToastError(f, err, "Failed to generate IP range for management network")
		}
	}

	if coreSubnet != "" {
		coreRange, err = h.newHostIPs(coreSubnet)
		if err != nil {
			return ToastError(f, err, "Failed to generate IP range for core network")
		}
	}

	type hostStruct struct {
		Host   string
		MgmtIp string
		CoreIp string
	}
	hosts := make([]hostStruct, len(hostArr))

	for i, host := range hostArr {
		hosts[i] = hostStruct{
			Host:   host,
			MgmtIp: mgmtRange[i],
			CoreIp: coreRange[i],
		}
	}

	return f.Render("fragments/hostAddModalInterfaces", fiber.Map{
		"Hosts": hosts,
	}, "")

}

func (h *Handler) userTable(f *fiber.Ctx) error {
	users, err := h.DB.GetUsers()
	if err != nil {
		return ToastError(f, err, "Failed to load users")
	}

	return f.Render("fragments/userTable", fiber.Map{
		"Users": users,
	}, "")
}

func (h *Handler) floorplanAddRack(f *fiber.Ctx) error {
	return f.Render("fragments/floorplanAddRack", fiber.Map{}, "")
}
