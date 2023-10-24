package frontend

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
	"github.com/ubccr/grendel/tors"
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

	return f.Render("fragments/host/form", fiber.Map{
		"Host":       host[0],
		"BootImages": h.getBootImages(),
		"Firmwares":  h.getFirmware(),
		"Interfaces": Interfaces,
	}, "")
}

func (h *Handler) rackTable(f *fiber.Ctx) error {
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

	return f.Render("fragments/rack/table", fiber.Map{
		"u":     u,
		"Hosts": filtered,
		"Rack":  rack,
	}, "")
}

func (h *Handler) rackAddModal(f *fiber.Ctx) error {
	return f.Render("fragments/rack/add/modal", fiber.Map{
		"Rack":       f.Params("rack"),
		"Firmwares":  h.getFirmware(),
		"BootImages": h.getBootImages(),
		"RackUs":     f.Query("hosts"), // rename me
	}, "")
}

type hostIfaceStruct struct {
	Port string
	MAC  string
	IP   string
}
type hostStruct struct {
	Name       string
	Interfaces []hostIfaceStruct
}
type interfaceStruct struct {
	Domain string
	Name   string
	BMC    string
	VLAN   string
	MTU    string
}
type RackAddFormStruct struct {
	Hosts      []hostStruct
	Interfaces []interfaceStruct
}

func (h *Handler) rackAddTable(f *fiber.Ctx) error {

	prefix := f.FormValue("Prefix")
	rack := f.Params("rack")
	rackUs := f.FormValue("rackUs")
	hostTable := f.FormValue("hostTable", "")
	ifaceCount, err := strconv.Atoi(f.FormValue("IfaceCount"))
	if err != nil {
		return ToastError(f, err, "Invalid interface count")
	}
	uArr := strings.Split(rackUs, ",")

	// Get list of switches for switch datalist
	dbHosts, err := h.DB.Hosts()
	if err != nil {
		return ToastError(f, err, "Failed to load hosts")
	}
	switches := dbHosts.FilterPrefix("swe")
	switchNames := make([]string, len(switches))
	for i, sw := range switches {
		switchNames[i] = sw.Name
	}

	// Hosts
	var hosts RackAddFormStruct

	if hostTable != "" {
		err := json.Unmarshal([]byte(hostTable), &hosts)
		if err != nil {
			return ToastError(f, err, "Failed to Unmarshal the host table")
		}
	} else {
		// Init new form
		hosts = RackAddFormStruct{
			Hosts:      make([]hostStruct, len(uArr)),
			Interfaces: make([]interfaceStruct, ifaceCount),
		}
		for i := 0; i < len(uArr); i++ {
			hosts.Hosts[i] = hostStruct{
				Name:       fmt.Sprintf("%s-%s-%s", prefix, rack, uArr[i]),
				Interfaces: make([]hostIfaceStruct, ifaceCount),
			}
			// Set first port (BMC) to same number as rack u
			hosts.Hosts[i].Interfaces[0].Port = uArr[i]
		}

		// TODO: this should be configurable
		// first iface (usually bmc)
		hosts.Interfaces[0] = interfaceStruct{
			Domain: "mgmt.ccr.buffalo.edu",
			Name:   "",
			BMC:    "true",
			VLAN:   "",
			MTU:    "1500",
		}

	}

	// Generate IP ranges from subnet
	interfaceIpArr := make([][]string, ifaceCount)
	for i := 0; i < ifaceCount; i++ {
		interfaceIpArr[i] = make([]string, len(hosts.Hosts))
		subnet := f.FormValue(fmt.Sprintf("subnet:%d", i))
		if subnet != "" {
			ipArr, err := h.newHostIPs(subnet)
			if err != nil {
				return ToastError(f, err, "Failed to generate IP range")
			}
			interfaceIpArr[i] = ipArr
		}
	}

	// Query MAC addresses
	interfaceMacArr := make([]tors.MACTable, ifaceCount)
	for i := 0; i < ifaceCount; i++ {
		sw := f.FormValue(fmt.Sprintf("switch:%d", i), "")
		if sw != "" && hosts.Hosts[0].Interfaces[i].MAC == "" {
			macTable, err := h.getMacAddress(sw)
			if err != nil {
				return ToastError(f, err, "Failed to get MAC address table")
			}
			interfaceMacArr[i] = macTable
		} else {
			interfaceMacArr[i] = nil
		}
	}

	// Map IP and MAC to hosts array
	for i, host := range hosts.Hosts {
		// update prefix if changed
		// possible bug if prefix is individually changed by user
		hostNameArr := strings.Split(host.Name, "-")
		if hostNameArr[0] != prefix {
			hosts.Hosts[i].Name = strings.Replace(host.Name, hostNameArr[0], prefix, 1)
		}
		for x, iface := range host.Interfaces {
			hosts.Hosts[i].Interfaces[x].IP = interfaceIpArr[x][i]
			if iface.Port != "" {
				port, err := strconv.Atoi(iface.Port)
				if err != nil {
					return ToastError(f, err, "Invalid port number")
				}
				MAC := interfaceMacArr[x].Port(port)
				if len(MAC) != 0 {
					hosts.Hosts[i].Interfaces[x].MAC = MAC[0].MAC.String()
				}
			}
		}
	}

	return f.Render("fragments/rack/add/table", fiber.Map{
		"Hosts":      hosts,
		"Switches":   switchNames,
		"IfaceCount": ifaceCount,
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
	return f.Render("fragments/floorplan/modal", fiber.Map{
		"Firmware":  h.getFirmware(),
		"BootImage": h.getBootImages(),
	}, "")
}

func (h *Handler) interfaces(f *fiber.Ctx) error {
	id := f.Query("ID", "0")

	return f.Render("fragments/interfaces", fiber.Map{
		"ID": id,
	}, "")
}
