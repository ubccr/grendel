// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
	"github.com/ubccr/grendel/internal/bmc"
	"github.com/ubccr/grendel/internal/firmware"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

func (h *Handler) LoginUser(f *fiber.Ctx) error {
	user := f.FormValue("username")
	pass := f.FormValue("password")

	val, role, err := h.DB.VerifyUser(user, pass)
	if err != nil || !val {
		msg := "Internal server error"
		if err.Error() == "crypto/bcrypt: hashedPassword is not the hash of the given password" {
			msg = "Failed to login: Password is incorrect"
		} else if err.Error() == "not found" {
			msg = "Failed to login: Username not found."
		}
		return ToastError(f, err, msg)
	}
	log.Debugf("User %s authenticated with role: %s", user, role)

	sess, err := h.Store.Get(f)
	if err != nil {
		return ToastError(f, err, "Failed to create session")
	}

	sess.Set("authenticated", true)
	sess.Set("user", user)
	sess.Set("role", role)

	err = sess.Save()
	if err != nil {
		return ToastError(f, err, "Failed to save session")
	}

	f.Response().Header.Add("HX-Redirect", "/")
	return ToastSuccess(f, "Successfully logged in", ``)
}

func (h *Handler) LogoutUser(f *fiber.Ctx) error {
	sess, _ := h.Store.Get(f)
	user := sess.Get("user")
	sess.Destroy()

	log.Debugf("User %s logged out", user)

	f.Response().Header.Add("HX-Redirect", "/")
	return ToastSuccess(f, "Successfully logged out", ``)
}

func (h *Handler) RegisterUser(f *fiber.Ctx) error {
	u := f.FormValue("username")
	p := f.FormValue("password")
	p2 := f.FormValue("password2")

	msg := "Successfully registered user"

	su := strings.ToValidUTF8(strings.TrimSpace(u), "")
	sp := strings.ToValidUTF8(strings.TrimSpace(p), "")

	if len(u) < 3 {
		return ToastError(f, nil, "Failed to register: Username must be at least 3 characters")
	} else if u != su {
		return ToastError(f, nil, "Failed to register: Username must not contain spaces or unicode characters")
	} else if p != p2 {
		return ToastError(f, nil, "Failed to register: Passwords do not match")
	} else if len(p) < 8 {
		return ToastError(f, nil, "Failed to register: Password must be at least 8 characters")
	} else if !strings.ContainsAny(sp, "abcdefghijklmnopqrstuvwxyz") {
		return ToastError(f, nil, "Failed to register: Password must contain at least one lowercase letter")
	} else if !strings.ContainsAny(sp, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return ToastError(f, nil, "Failed to register: Password must contain at least one uppercase letter")
	} else if !strings.ContainsAny(sp, "0123456789") {
		return ToastError(f, nil, "Failed to register: Password must contain at least one number")
	} else if !strings.ContainsAny(sp, "!@#$%^&*()") {
		return ToastError(f, nil, "Failed to register: Password must contain at least one special character")
	} else if p != sp {
		return ToastError(f, nil, "Failed to register: Password must not contain spaces or unicode characters")
	}

	role, err := h.DB.StoreUser(su, sp)
	if err != nil {
		if err.Error() == fmt.Sprintf("User %s already exists", su) {
			msg = "Failed to register: Username already exists"
		} else {
			msg = "Failed to register user"
		}
		return ToastError(f, err, msg)
	}

	log.Debugf("New user: %s registered", su)

	sess, err := h.Store.Get(f)
	if err != nil {
		return ToastError(f, err, "Failed to create session")
	}

	sess.Set("authenticated", true)
	sess.Set("user", su)
	sess.Set("role", role)

	err = sess.Save()
	if err != nil {
		return ToastError(f, err, "Failed to save session")
	}

	h.writeEvent("info", f, fmt.Sprintf("New user: %s registered.", su))
	f.Set("HX-Redirect", "/")
	return ToastSuccess(f, msg, ``)
}

func (h *Handler) deleteUser(f *fiber.Ctx) error {
	user := f.Params("username")

	err := h.DB.DeleteUser(user)
	if err != nil {
		return ToastError(f, err, "Failed to delete user")
	}

	log.Debugf("User %s deleted", user)

	h.writeEvent("info", f, fmt.Sprintf("Deleted user: %s", user))
	return ToastSuccess(f, "Successfully deleted user", `, "refresh": ""`)
}

type FormData struct {
	ID         string `form:"ID"`
	Name       string `form:"Name"`
	Provision  string `form:"Provision"`
	Firmware   string `form:"Firmware"`
	BootImage  string `form:"BootImage"`
	Tags       string `form:"Tags"`
	Interfaces string `form:"Interfaces"`
	Bonds      string `form:"Bonds"`
}
type InterfacesString struct {
	FQDN string `json:"Fqdn"`
	MAC  string `json:"Mac"`
	IP   string `json:"Ip"`
	Name string `json:"Ifname"`
	BMC  string `json:"bmc"`
	VLAN string `json:"Vlan"`
	MTU  string `json:"Mtu"`
}
type BondsString struct {
	InterfacesString
	Peers string `json:"Peers"`
}

func (h *Handler) EditHost(f *fiber.Ctx) error {
	formHost := new(FormData)

	err := f.BodyParser(formHost)
	if err != nil {
		return ToastError(f, err, "Failed to bind type to request body")
	}

	id, _ := ksuid.Parse(formHost.ID)

	provision, err := strconv.ParseBool(formHost.Provision)
	if err != nil {
		return ToastError(f, err, "Failed to parse provision boolean")
	}

	var ifaces []InterfacesString
	json.Unmarshal([]byte(formHost.Interfaces), &ifaces)
	var bondsStr []BondsString
	json.Unmarshal([]byte(formHost.Bonds), &bondsStr)

	var interfaces []*model.NetInterface
	var bonds []*model.Bond

	for i, iface := range ifaces {
		mac, _ := net.ParseMAC(iface.MAC)
		ip, _ := netip.ParsePrefix(iface.IP)
		bmc, err := strconv.ParseBool(iface.BMC)
		if err != nil {
			return ToastError(f, err, fmt.Sprintf("Failed to parse BMC boolean on interface %d", i))
		}
		mtu, err := strconv.Atoi(iface.MTU)
		if err != nil {
			return ToastError(f, err, fmt.Sprintf("Failed to parse MTU on interface %d", i))
		}

		interfaces = append(interfaces, &model.NetInterface{
			Name: iface.Name,
			FQDN: iface.FQDN,
			MAC:  mac,
			IP:   ip,
			BMC:  bmc,
			VLAN: iface.VLAN,
			MTU:  uint16(mtu),
		})
	}
	for i, bond := range bondsStr {
		mac, _ := net.ParseMAC(bond.MAC)
		ip, _ := netip.ParsePrefix(bond.IP)
		bmc, err := strconv.ParseBool(bond.BMC)
		if err != nil {
			return ToastError(f, err, fmt.Sprintf("Failed to parse BMC boolean on interface %d", i))
		}
		mtu, err := strconv.Atoi(bond.MTU)
		if err != nil {
			return ToastError(f, err, fmt.Sprintf("Failed to parse MTU on interface %d", i))
		}

		bonds = append(bonds, &model.Bond{
			NetInterface: model.NetInterface{
				Name: bond.Name,
				FQDN: bond.FQDN,
				MAC:  mac,
				IP:   ip,
				BMC:  bmc,
				VLAN: bond.VLAN,
				MTU:  uint16(mtu),
			},
			Peers: strings.Split(bond.Peers, ","),
		})
	}

	newHost := model.Host{
		UID:        id,
		Name:       formHost.Name,
		Provision:  provision,
		Firmware:   firmware.NewFromString(formHost.Firmware),
		BootImage:  formHost.BootImage,
		Tags:       strings.Split(formHost.Tags, ","),
		Interfaces: interfaces,
		Bonds:      bonds,
	}

	err = h.DB.StoreHost(&newHost)
	if err != nil {
		return ToastError(f, err, "Failed to update host")
	}

	h.writeEvent("info", f, fmt.Sprintf("Edited or added host %s", newHost.Name))
	return ToastSuccess(f, "Successfully updated host", `, "refresh": ""`)
}

type inventoryHostData struct {
	SerialNumber string `form:"SerialNumber"`
	AssetNumber  string `form:"AssetNumber"`
	Manufacturer string `form:"Manufacturer"`
	Model        string `form:"Model"`
	Tags         string `form:"Tags"`
	PONumber     string `form:"PONumber"`
}

func (h *Handler) AddInventoryHost(f *fiber.Ctx) error {
	form := new(inventoryHostData)

	err := f.BodyParser(form)
	if err != nil {
		return ToastError(f, err, "Failed to bind type to request body")
	}

	name := fmt.Sprintf("inv-%s", form.SerialNumber)
	tags := []string{}

	serial := fmt.Sprintf("serial_number:%s", form.SerialNumber)
	tags = append(tags, serial)

	if form.AssetNumber != "" {
		t := fmt.Sprintf("asset_number:%s", form.AssetNumber)
		tags = append(tags, t)
	}

	if form.Manufacturer != "" {
		t := fmt.Sprintf("manufacturer:%s", form.Manufacturer)
		tags = append(tags, t)
	}
	if form.Model != "" {
		t := fmt.Sprintf("model:%s", form.Model)
		tags = append(tags, t)
	}

	if form.PONumber != "" {
		t := fmt.Sprintf("po_number:%s", form.PONumber)
		tags = append(tags, t)
	}

	if form.Tags != "" {
		tags = append(tags, strings.Split(form.Tags, ",")...)
	}

	newHost := model.Host{
		Name: name,
		Tags: tags,
	}

	ns, err := nodeset.NewNodeSet(name)
	if err != nil {
		f.Status(400)
		return ToastError(f, err, "Failed to check for duplicates")
	}
	hl, err := h.DB.FindHosts(ns)
	if err != nil {
		f.Status(400)
		return ToastError(f, err, "Failed to lookup duplicates")
	}

	if len(hl) != 0 {
		f.Status(400)
		ns, _ := hl.ToNodeSet()
		return ToastError(f, err, fmt.Sprintf("Failed to add inventory host. Duplicate tag entry detected: %s", ns))
	}

	hlt, _ := h.DB.FindTags([]string{serial})

	if hlt != nil && hlt.Len() != 0 {
		f.Status(400)
		return ToastError(f, err, fmt.Sprintf("Failed to add inventory host. Duplicate tag entry detected: %s", hlt))
	}

	err = h.DB.StoreHost(&newHost)
	if err != nil {
		f.Status(400)
		return ToastError(f, err, "Failed to update host")
	}

	h.writeEvent("info", f, fmt.Sprintf("Added new Inventory host %s", newHost.Name))
	return ToastSuccess(f, "Successfully added inventory host", `, "refresh": ""`)
}

func (h *Handler) DeleteHost(f *fiber.Ctx) error {
	hosts := f.FormValue("hosts")
	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	err = h.DB.DeleteHosts(ns)
	if err != nil {
		return ToastError(f, err, "Failed to delete host(s)")
	}

	h.writeEvent("info", f, fmt.Sprintf("Deleted host(s): %s", ns))
	return ToastSuccess(f, "Successfully deleted host(s)", `, "refresh":""`)
}

func (h *Handler) importHost(f *fiber.Ctx) error {
	s := f.FormValue("json")
	i := model.HostList{}

	err := json.Unmarshal([]byte(s), &i)
	if err != nil {
		return ToastError(f, err, fmt.Sprintf("Failed to unmarshal json: %s", err))
	}

	// Generate new ksuid to ensure no conflicts
	for _, h := range i {
		h.UID = ksuid.New()
	}

	err = h.DB.StoreHosts(i)
	if err != nil {
		return ToastError(f, err, "Failed to import host")
	}

	n, _ := i.ToNodeSet()

	h.writeEvent("info", f, fmt.Sprintf("imported host(s) %s", n))
	return ToastSuccess(f, fmt.Sprintf("Successfully imported host(s) %s", n), `, "refresh": ""`)
}

func (h *Handler) bmcPowerCycle(f *fiber.Ctx) error {
	powerOption := f.FormValue("power-option")
	bootOption := f.FormValue("boot-override-option")
	hosts := f.FormValue("hosts")

	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return ToastError(f, err, "Failed to find nodes")
	}

	job := bmc.NewJob()

	var jobMessages []bmc.JobMessage

	switch powerOption {
	case "power-cycle":
		jobMessages, err = job.PowerCycle(hostList, bootOption)
	case "power-on":
		jobMessages, err = job.PowerOn(hostList, bootOption)
	case "power-off":
		jobMessages, err = job.PowerOff(hostList)
	}
	if err != nil {
		return ToastError(f, err, "Failed to run power job")
	}

	h.writeJobEvent(f, fmt.Sprintf("Submitted host %s job", powerOption), jobMessages)

	return ToastSuccess(f, "Successfully submitted power job on node(s)", ``)
}

func (h *Handler) bmcPowerCycleBmc(f *fiber.Ctx) error {
	hosts := f.FormValue("hosts")

	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return ToastError(f, err, "Failed to find nodes")
	}

	job := bmc.NewJob()
	jobMessages, err := job.PowerCycleBmc(hostList)
	if err != nil {
		return ToastError(f, err, "Failed to run power cycle bmc job")
	}
	h.writeJobEvent(f, "Submitted power cycle BMC job", jobMessages)

	return ToastSuccess(f, "Successfully submitted power cycle bmc job on node(s)", ``)
}

func (h *Handler) bmcClearSel(f *fiber.Ctx) error {
	hosts := f.FormValue("hosts")

	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return ToastError(f, err, "Failed to find nodes")
	}

	job := bmc.NewJob()

	jobMessages, err := job.ClearSel(hostList)
	if err != nil {
		return ToastError(f, err, "Failed to run clear sel job")
	}
	h.writeJobEvent(f, "Submitted clear SEL job", jobMessages)

	return ToastSuccess(f, "Successfully submitted clear sel job on node(s)", ``)
}

func (h *Handler) provisionHosts(f *fiber.Ctx) error {
	hosts := f.FormValue("hosts")
	provision := f.FormValue("provision")

	p, err := strconv.ParseBool(provision)
	if err != nil {
		return ToastError(f, err, "Failed to parse provision boolean")
	}

	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	err = h.DB.ProvisionHosts(ns, p)
	if err != nil {
		return ToastError(f, err, "Failed to provision host(s)")
	}

	h.writeEvent("info", f, fmt.Sprintf("Provisioned host(s): %s", ns))
	return ToastSuccess(f, "Successfully provisioned host(s)", `, "refresh": ""`)
}

func (h *Handler) tagHosts(f *fiber.Ctx) error {
	hosts := f.FormValue("hosts")
	tags := f.FormValue("tags")
	action := f.FormValue("action")

	t := strings.Split(tags, ",")

	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	if action == "remove" {
		err = h.DB.UntagHosts(ns, t)
	} else {
		err = h.DB.TagHosts(ns, t)
	}
	if err != nil {
		return ToastError(f, err, "Failed to update tags on host(s)")
	}

	h.writeEvent("info", f, fmt.Sprintf("%sed tags to %s", action, ns))
	return ToastSuccess(f, "Successfully updated tags on host(s)", `, "refresh": ""`)
}

func (h *Handler) imageHosts(f *fiber.Ctx) error {
	hosts := f.FormValue("hosts")
	image := f.FormValue("image")

	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	err = h.DB.SetBootImage(ns, image)
	if err != nil {
		return ToastError(f, err, "Failed to update boot image on host(s)")
	}

	h.writeEvent("info", f, fmt.Sprintf("Updated boot image to %s on %s", image, ns))
	return ToastSuccess(f, "Successfully updated boot image on host(s)", `, "refresh": ""`)
}

func (h *Handler) exportHosts(f *fiber.Ctx) error {
	hosts := f.Params("hosts")
	filename := f.Query("filename")

	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return ToastError(f, err, "Failed to find host(s)")
	}

	o, err := json.MarshalIndent(hostList, "", "  ")
	if err != nil {
		return ToastError(f, err, "Failed to marshal host json")
	}

	if filename != "" {
		f.Set("HX-Redirect", fmt.Sprintf("/api/hosts/export/%s?filename=%s", hosts, filename))
		f.Set("Content-Type", "application/force-download")
		f.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		ToastSuccess(f, "Download starting shortly...", "")
	} else {
		f.Response().Header.Set("HX-Trigger", "openExportModal")
	}
	return f.SendString(string(o))
}

type templateDataHosts struct {
	Manufacturer string
	Model        string
	SerialNumber string
	AssetNumber  string
	PONumber     string
	Host         *model.Host
	System       bmc.System
}

type templateData struct {
	Hosts []templateDataHosts
	Date  string
}

func (h *Handler) exportInventory(f *fiber.Ctx) error {
	hosts := f.Params("hosts")
	filename := f.Query("filename")
	templateString := f.Query("template")

	tmpl, err := template.New("test").Parse(templateString)
	if err != nil {
		return ToastError(f, err, "Failed to parse template")
	}

	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return ToastError(f, err, "Failed to find nodes")
	}

	job := bmc.NewJob()
	jobStatuses, err := job.BmcStatus(hostList)
	if err != nil {
		return err
	}

	td := templateData{
		Date: time.Now().Format(time.DateOnly),
	}

	queryErrors := "\n\nWARNING:\n"
	for _, host := range hostList {
		assetNumbers := []string{}
		for _, tag := range host.Tags {
			if !strings.Contains(tag, "asset_number") {
				continue
			}
			t := strings.Split(tag, ":")
			if len(t) < 2 {
				continue
			}
			assetNumbers = append(assetNumbers, t[1])
		}

		if len(assetNumbers) == 0 {
			assetNumbers = append(assetNumbers, "")
		}

		for _, assetNumber := range assetNumbers {
			tdh := templateDataHosts{
				Host:        host,
				System:      bmc.System{},
				AssetNumber: assetNumber,
			}
			for _, jobStatus := range jobStatuses {
				if jobStatus.Name != host.Name {
					continue
				}
				tdh.System = jobStatus
				tdh.Manufacturer = jobStatus.Manufacturer
				tdh.Model = jobStatus.Model
				tdh.SerialNumber = jobStatus.SerialNumber
				break
			}
			for _, tag := range host.Tags {
				t := strings.Split(tag, ":")
				if len(t) < 2 {
					continue
				}

				switch t[0] {
				case "serial_number":
					if tdh.SerialNumber != "" && tdh.SerialNumber != t[1] {
						queryErrors += fmt.Sprintf("Host: %s, Tag: %s does not match queried value of: %s\n", host.Name, t, tdh.SerialNumber)
					}
					tdh.SerialNumber = t[1]
				case "po_number":
					tdh.PONumber = t[1]
				case "model":
					if tdh.Model != "" && tdh.Model != t[1] {
						queryErrors += fmt.Sprintf("Host: %s, Tag: %s does not match queried value of: %s\n", host.Name, t, tdh.Model)
					}
					tdh.Model = t[1]
				case "manufacturer":
					if tdh.Manufacturer != "" && tdh.Manufacturer != t[1] {
						queryErrors += fmt.Sprintf("Host: %s, Tag: %s does not match queried value of: %s\n", host.Name, t, tdh.Manufacturer)
					}
					tdh.Manufacturer = t[1]
				}
			}

			td.Hosts = append(td.Hosts, tdh)
		}
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, td)
	if err != nil {
		return ToastError(f, err, "Failed to execute template")
	}

	if queryErrors != "\n\nWARNING:\n" {
		buf.WriteString(queryErrors)
	}

	if filename != "" {
		f.Set("HX-Redirect", fmt.Sprintf("/api/hosts/inventory/%s?template=%s&filename=%s", hosts, url.QueryEscape(templateString), filename))
		f.Set("Content-Type", "application/force-download")
		f.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		ToastSuccess(f, "Download starting shortly...", "")
	} else {
		f.Response().Header.Set("HX-Trigger", "openExportModal")
	}

	return f.SendString(buf.String())
}

func (h *Handler) bmcConfigureAuto(f *fiber.Ctx) error {
	hosts := f.FormValue("hosts")
	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	hostList, _ := h.DB.FindHosts(ns)
	if err != nil {
		return ToastError(f, err, "Failed to find nodes")
	}

	job := bmc.NewJob()

	jobMessages, err := job.BmcAutoConfigure(hostList)
	if err != nil {
		return ToastError(f, err, "Failed to run auto config job")
	}
	h.writeJobEvent(f, "Submitted auto configure job", jobMessages)

	return ToastSuccess(f, "Successfully sent Auto Configure to node(s)", ``)
}
func (h *Handler) bmcConfigureImport(f *fiber.Ctx) error {
	file := f.FormValue("File")
	shutdownType := f.FormValue("shutdownType")
	if file == "" {
		return ToastError(f, nil, "No file specified")
	}
	if shutdownType == "" {
		return ToastError(f, nil, "No Shutdown Type specified")
	}

	hosts := f.FormValue("hosts")
	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	hostList, _ := h.DB.FindHosts(ns)
	if err != nil {
		return ToastError(f, err, "Failed to find nodes")
	}

	job := bmc.NewJob()
	jobMessages, err := job.BmcImportConfiguration(hostList, shutdownType, file)
	if err != nil {
		return ToastError(f, err, "Failed to run import config job")
	}
	h.writeJobEvent(f, "Submitted import system configuration job", jobMessages)

	return ToastSuccess(f, "Successfully sent system config to node(s)", ``)
}

func (h *Handler) Search(f *fiber.Ctx) error {
	search := f.FormValue("search")

	f.Response().Header.Add("HX-Redirect", fmt.Sprintf("/host/%s", search))
	return nil
}

func (h *Handler) usersPost(f *fiber.Ctx) error {
	users := f.FormValue("Usernames")
	role := f.FormValue("Role")
	userList := strings.Split(users, ",")

	for _, user := range userList {
		err := h.DB.UpdateUserRole(user, role)
		if err != nil {
			return ToastError(f, err, "Failed to update user: "+user)
		}
	}

	h.writeEvent("info", f, fmt.Sprintf("Changed user(s): %s to Role: %s", users, role))
	return ToastSuccess(f, "Successfully updated user(s)", `, "refresh": ""`)
}

func (h *Handler) bulkHostAdd(f *fiber.Ctx) error {
	type formStruct struct {
		Provision string `form:"Provision"`
		Firmware  string `form:"Firmware"`
		BootImage string `form:"BootImage"`
		Tags      string `form:"Tags"`
	}

	var formData formStruct
	err := f.BodyParser(&formData)
	if err != nil {
		return ToastError(f, err, "Failed to bind form data")
	}

	var hostTableForm RackAddFormStruct
	err = json.Unmarshal([]byte(f.FormValue("hostTable")), &hostTableForm)
	if err != nil {
		return ToastError(f, err, "Failed to Unmarshal the host table")
	}

	var newHosts []*model.Host
	provision, err := strconv.ParseBool(formData.Provision)
	if err != nil {
		return ToastError(f, err, "Failed to parse provision boolean")
	}

	for _, host := range hostTableForm.Hosts {
		var newInterface []*model.NetInterface
		for i, iface := range host.Interfaces {
			hostName := host.Name
			hostNameArr := strings.Split(hostName, "-")
			if len(hostNameArr) < 1 {
				return ToastError(f, err, "Failed to parse host name")
			}
			if hostTableForm.Interfaces[i].BMC == "true" && hostNameArr[0] != "swe" && hostNameArr[0] != "swi" {
				hostName = strings.Replace(hostName, hostNameArr[0], "bmc", 1)
			}

			fqdn := fmt.Sprintf("%s.%s", hostName, hostTableForm.Interfaces[i].Domain)
			mac, _ := net.ParseMAC(iface.MAC)
			ip, _ := netip.ParsePrefix(iface.IP)
			bmc, err := strconv.ParseBool(hostTableForm.Interfaces[i].BMC)
			if err != nil {
				return ToastError(f, err, "Failed to parse BMC boolean")
			}
			mtu, err := strconv.Atoi(hostTableForm.Interfaces[i].MTU)
			if err != nil {
				return ToastError(f, err, "Failed to parse MTU")
			}

			newInterface = append(newInterface, &model.NetInterface{
				Name: hostTableForm.Interfaces[i].Name,
				FQDN: fqdn,
				MAC:  mac,
				IP:   ip,
				BMC:  bmc,
				VLAN: hostTableForm.Interfaces[i].VLAN,
				MTU:  uint16(mtu),
			})
		}

		newHosts = append(newHosts, &model.Host{
			UID:        ksuid.New(),
			Name:       host.Name,
			Provision:  provision,
			BootImage:  formData.BootImage,
			Firmware:   firmware.NewFromString(formData.Firmware),
			Tags:       strings.Split(formData.Tags, ","),
			Interfaces: newInterface,
		})
	}

	err = h.DB.StoreHosts(newHosts)
	if err != nil {
		return ToastError(f, err, "Failed to add host(s)")
	}
	var jobMessages []bmc.JobMessage
	for _, host := range newHosts {
		jobMessages = append(jobMessages, bmc.JobMessage{
			Status: "success",
			Host:   host.Name,
			Msg:    "Successfully added host",
		})
	}
	h.writeJobEvent(f, "Submitted bulk add hosts job", jobMessages)

	return ToastSuccess(f, "Successfully added host(s)", `, "closeModal": "", "refresh": ""`)
}
