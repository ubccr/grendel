package frontend

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/bmc"
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
	type BondStrings struct {
		FQDN  string
		MAC   string
		IP    string
		Name  string
		BMC   string
		VLAN  string
		MTU   string
		Peers []string
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
	Bonds := make([]BondStrings, len(host[0].Bonds))

	for i, bond := range host[0].Bonds {
		Bonds[i].FQDN = bond.FQDN
		Bonds[i].MAC = bond.MAC.String()
		Bonds[i].IP = bond.IP.String()
		Bonds[i].Name = bond.Name
		Bonds[i].BMC = strconv.FormatBool(bond.BMC)
		Bonds[i].VLAN = bond.VLAN
		Bonds[i].MTU = strconv.FormatUint(uint64(bond.MTU), 10)
		Bonds[i].Peers = bond.Peers
	}

	return f.Render("fragments/host/form", fiber.Map{
		"Host":       host[0],
		"BootImages": h.getBootImages(),
		"Firmwares":  h.getFirmware(),
		"Interfaces": Interfaces,
		"Bonds":      Bonds,
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

	type hostArrStruct struct {
		U     string
		Hosts model.HostList
	}
	hostArr := make([]hostArrStruct, 0)

	viper.SetDefault("frontend.rack_min", 3)
	viper.SetDefault("frontend.rack_max", 42)
	min := viper.GetInt("frontend.rack_min")
	max := viper.GetInt("frontend.rack_max")

	for i := max; i >= min; i-- {
		u := fmt.Sprintf("%02d", i)
		h := model.HostList{}

		for _, v := range hosts {
			if v.HostType() == "power" && !v.HasAnyTags("1u", "2u") {
				continue
			}
			nameArr := strings.Split(v.Name, "-")
			if len(nameArr) < 2 {
				log.Debugf("Invalid host name: %s", v.Name)
				continue
			}
			if nameArr[2] == u {
				h = append(h, v)
			}
		}

		hostArr = append(hostArr, hostArrStruct{
			U:     u,
			Hosts: h,
		})
	}

	return f.Render("fragments/rack/table", fiber.Map{
		"Hosts": hostArr,
		"Rack":  rack,
	}, "")
}

func (h *Handler) actions(f *fiber.Ctx) error {
	hosts := f.FormValue("hosts")
	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Invalid host set")
	}

	nodeset := ns.String()
	return f.Render("fragments/actions", fiber.Map{
		"Hosts":                    nodeset,
		"BmcSystem":                bmc.System{},
		"ExportCSVDefaultTemplate": viper.GetString("frontend.export_csv_default_template"),
		"BootImages":               h.getBootImages(),
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

	// Defaults
	defaultIface := interfaceStruct{
		Domain: viper.GetString("frontend.other_ifaces.interface_domain"),
		Name:   viper.GetString("frontend.other_ifaces.interface_name"),
		BMC:    viper.GetString("frontend.other_ifaces.interface_bmc"),
		VLAN:   viper.GetString("frontend.other_ifaces.interface_vlan"),
		MTU:    viper.GetString("frontend.other_ifaces.interface_mtu"),
	}
	defaultFirstIface := interfaceStruct{
		Domain: viper.GetString("frontend.first_iface.interface_domain"),
		Name:   viper.GetString("frontend.first_iface.interface_name"),
		BMC:    viper.GetString("frontend.first_iface.interface_bmc"),
		VLAN:   viper.GetString("frontend.first_iface.interface_vlan"),
		MTU:    viper.GetString("frontend.first_iface.interface_mtu"),
	}

	// Hosts
	var hosts RackAddFormStruct

	if hostTable != "" {
		err := json.Unmarshal([]byte(hostTable), &hosts)
		if err != nil {
			return ToastError(f, err, "Failed to Unmarshal the host table")
		}

		// init new iface if needed
		if len(hosts.Interfaces) < ifaceCount {
			hosts.Interfaces = append(hosts.Interfaces, defaultIface)
			for h := range hosts.Hosts {
				hosts.Hosts[h].Interfaces = append(hosts.Hosts[h].Interfaces, hostIfaceStruct{
					Port: "",
					MAC:  "",
					IP:   "",
				})
			}
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
			if viper.GetBool("frontend.first_iface.auto_mapping") {
				hosts.Hosts[i].Interfaces[0].Port = uArr[i]
			}
		}

		// first iface (usually bmc)
		hosts.Interfaces[0] = defaultFirstIface

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

func (h *Handler) nodesTable(f *fiber.Ctx) error {
	hosts, err := h.DB.Hosts()
	if err != nil {
		return ToastError(f, err, "Error getting hosts from DB")
	}
	qp := f.Queries()

	matches := []string{}

	// Filter Regex:
	nameRegex, err := regexp.Compile(qp["Name"])
	if err != nil {
		return ToastError(f, err, "Invalid Name Regex")
	}
	provisionRegex, err := regexp.Compile(qp["Provision"])
	if err != nil {
		return ToastError(f, err, "Invalid Boot Image Regex")
	}
	firmwareRegex, err := regexp.Compile(qp["Firmware"])
	if err != nil {
		return ToastError(f, err, "Invalid Boot Image Regex")
	}
	bootImageRegex, err := regexp.Compile(qp["BootImage"])
	if err != nil {
		return ToastError(f, err, "Invalid Boot Image Regex")
	}
	tagsRegex, err := regexp.Compile(qp["Tags"])
	if err != nil {
		return ToastError(f, err, "Invalid Tag Regex")
	}

	for _, h := range hosts {
		nameMatch := false
		provisionMatch := false
		firmwareMatch := false
		bootImageMatch := false
		tagsMatch := false

		if nameRegex.MatchString(h.Name) {
			nameMatch = true
		}
		if provisionRegex.MatchString(strconv.FormatBool(h.Provision)) {
			provisionMatch = true
		}
		if firmwareRegex.MatchString(h.Firmware.String()) {
			firmwareMatch = true
		}
		if bootImageRegex.MatchString(h.BootImage) {
			bootImageMatch = true
		}
		for _, tag := range h.Tags {
			if tagsRegex.MatchString(tag) {
				tagsMatch = true
				break
			}
		}

		if nameMatch && provisionMatch && firmwareMatch && bootImageMatch && tagsMatch {
			matches = append(matches, h.Name)
		}

	}

	ns, err := nodeset.NewNodeSet(strings.Join(matches, ","))
	if err != nil {
		return ToastError(f, err, "Error filtering hosts")
	}
	filtered, err := h.DB.FindHosts(ns)
	if err != nil {
		return ToastError(f, err, "Error filtering hosts")
	}
	sort.Slice(filtered, func(i, j int) bool {
		o := strings.Compare(filtered[i].Name, filtered[j].Name)

		return o == -1
	})
	allHosts, err := filtered.ToNodeSet()
	if err != nil {
		return ToastError(f, err, "Error filtering hosts")
	}

	// Pagination:
	pageSize, err := strconv.Atoi(qp["PageSize"])
	if err != nil {
		return ToastError(f, err, "Error calculating pagination")
	}
	filteredLen := len(filtered)
	numPages := filteredLen / pageSize
	if math.Remainder(float64(filteredLen), float64(pageSize)) != 0 {
		numPages++
	}

	currentPage := 1
	if qp["CurrentPage"] != "" {
		currentPage, err = strconv.Atoi(qp["CurrentPage"])
		if err != nil {
			return ToastError(f, err, "Error calculating pagination")
		}
	}

	start := (currentPage - 1) * pageSize

	if start > filteredLen {
		start = filteredLen
	}

	end := start + pageSize
	if end > filteredLen {
		end = filteredLen
	}

	filteredList := filtered[start:end]

	checkAll := false
	if qp["CheckAll"] == "on" {
		checkAll = true
	}

	return f.Render("fragments/nodes/table", fiber.Map{
		"Hosts":    filteredList,
		"AllHosts": allHosts,
		"CheckAll": checkAll,
		"Pagination": fiber.Map{
			"CurrentPage": currentPage,
			"PageSize":    pageSize,
			"NumPages":    numPages,
		},
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
		name := strings.Split(host.Name, "-")
		if len(name) < 2 {
			continue
		}
		rack := name[1]
		racks[rack] += 1
	}

	// Probably needs a rewrite for very large floorplans
	viper.SetDefault("frontend.rows_start", "f")
	viper.SetDefault("frontend.rows_end", "v")
	rStart := []rune(viper.GetString("frontend.rows_start"))
	rEnd := []rune(viper.GetString("frontend.rows_end"))

	rows := make([]string, 0)
	for i := rStart[0]; i <= rEnd[0]; i++ {
		rows = append(rows, fmt.Sprintf("%c", i))
	}

	viper.SetDefault("frontend.cols_start", 28)
	viper.SetDefault("frontend.cols_end", 5)
	cStart := viper.GetInt("frontend.cols_start")
	cEnd := viper.GetInt("frontend.cols_end")

	cols := make([]string, 0)
	for i := cStart; i >= cEnd; i-- {
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

func (h *Handler) bonds(f *fiber.Ctx) error {
	id := f.Query("ID", "0")

	return f.Render("fragments/bonds", fiber.Map{
		"ID": id,
	}, "")
}

func (h *Handler) events(f *fiber.Ctx) error {
	return f.Render("fragments/events/table", fiber.Map{
		"Events": h.Events,
	}, "")
}
