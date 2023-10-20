package frontend

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/nodeset"
)

func (h *Handler) hostForm(f *fiber.Ctx) error {
	reqHost, err := nodeset.NewNodeSet(f.Params("host"))
	if err != nil {
		return ToastError(f, fmt.Errorf("invalid host"), "Invalid host")
	}

	host, err := h.DB.FindHosts(reqHost)
	if err != nil || len(host) == 0 {
		return ToastError(f, err, "Failed to find host")
	}
	type IfaceStrings struct {
		FQDN string
		MAC  string
		IP   string
		Name string
		BMC  string
		VLAN string
		MTU  string
	}
	Interfaces := make([]IfaceStrings, len(host[0].Interfaces))

	for i, iface := range host[0].Interfaces {
		Interfaces[i].FQDN = iface.FQDN
		Interfaces[i].MAC = iface.MAC.String()
		Interfaces[i].IP = iface.IP.String()
		Interfaces[i].Name = iface.Name
		Interfaces[i].BMC = strconv.FormatBool(iface.BMC)
		Interfaces[i].VLAN = iface.VLAN
		Interfaces[i].MTU = strconv.FormatUint(uint64(iface.MTU), 10)
	}

	images, _ := h.DB.BootImages()
	bootImages := make([]string, 0)
	for _, i := range images {
		bootImages = append(bootImages, i.Name)
	}

	fw := make([]string, 0)
	for _, i := range firmware.BuildToStringMap {
		fw = append(fw, i)
	}

	return f.Render("fragments/host/form", fiber.Map{
		"Host":       host[0],
		"BootImages": bootImages,
		"Firmwares":  fw,
		"Interfaces": Interfaces,
	}, "")
}

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

func (h *Handler) usersTable(f *fiber.Ctx) error {
	users, err := h.DB.GetUsers()
	if err != nil {
		return ToastError(f, err, "Failed to load users")
	}

	return f.Render("fragments/users/table", fiber.Map{
		"Users": users,
	}, "")
}

func (h *Handler) floorplanTable(f *fiber.Ctx) error {
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
	return f.Render("fragments/floorplan/table", fiber.Map{
		"Rows":  rows,
		"Cols":  cols,
		"Racks": racks,
	}, "")
}

func (h *Handler) floorplanModal(f *fiber.Ctx) error {
	fw := make([]string, 0)
	for _, i := range firmware.BuildToStringMap {
		fw = append(fw, i)
	}

	images, _ := h.DB.BootImages()
	bootImages := make([]string, 0)
	for _, i := range images {
		bootImages = append(bootImages, i.Name)
	}

	return f.Render("fragments/floorplan/modal", fiber.Map{
		"Firmware":  fw,
		"BootImage": bootImages,
	}, "")
}
func (h *Handler) interfaces(f *fiber.Ctx) error {
	id := f.Query("ID", "0")

	return f.Render("fragments/interfaces", fiber.Map{
		"ID": id,
	}, "")
}
