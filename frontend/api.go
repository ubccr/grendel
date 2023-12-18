package frontend

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish/redfish"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
	"github.com/valyala/fasthttp"
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

	err := h.DB.StoreUser(su, sp)
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
	sess.Set("role", "disabled")

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

	var interfaces []*model.NetInterface

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

	newHost := model.Host{
		ID:         id,
		Name:       formHost.Name,
		Provision:  provision,
		Firmware:   firmware.NewFromString(formHost.Firmware),
		BootImage:  formHost.BootImage,
		Tags:       strings.Split(formHost.Tags, ","),
		Interfaces: interfaces,
	}

	err = h.DB.StoreHost(&newHost)
	if err != nil {
		return ToastError(f, err, "Failed to update host")
	}

	h.writeEvent("info", f, fmt.Sprintf("Edited or added host %s", newHost.Name))
	return ToastSuccess(f, "Successfully updated host", `, "refresh": ""`)
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

type RebootData struct {
	Host string
}

func (h *Handler) RebootHost(f *fiber.Ctx) error {
	delay := viper.GetInt("bmc.delay")
	fanout := viper.GetInt("bmc.fanout")

	bootOption := f.FormValue("boot-option")
	hosts := f.FormValue("hosts")
	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return ToastError(f, err, "Failed to find nodes")
	}

	runner := NewJobRunner(fanout)
	boot := redfish.Boot{
		BootSourceOverrideTarget:  redfish.NoneBootSourceOverrideTarget,
		BootSourceOverrideEnabled: redfish.OnceBootSourceOverrideEnabled,
	}
	if bootOption == "pxe" {
		boot.BootSourceOverrideTarget = redfish.PxeBootSourceOverrideTarget
	} else if bootOption == "bios-setup" {
		boot.BootSourceOverrideTarget = redfish.BiosSetupBootSourceOverrideTarget
	}

	for i, host := range hostList {
		ch := make(chan string)
		runner.RunReboot(host, ch, boot)
		if (i+1)%fanout == 0 {
			time.Sleep(time.Duration(delay) * time.Second)
		}
		output := strings.Split(<-ch, "|")

		if len(output) < 3 {
			return ToastError(f, nil, "Failed to run reboot job, index out of range")
		}
		h.writeEvent(output[0], f, fmt.Sprintf("%s: %s", output[1], output[2]))
	}
	runner.Wait()

	return ToastSuccess(f, "Successfully Rebooted node(s)", ``)
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
	}
	return f.SendString(string(o))
}

func (h *Handler) bmcConfigureAuto(f *fiber.Ctx) error {
	delay := viper.GetInt("bmc.delay")
	fanout := viper.GetInt("bmc.fanout")

	hosts := f.FormValue("hosts")
	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	hostList, _ := h.DB.FindHosts(ns)
	if err != nil {
		return ToastError(f, err, "Failed to find nodes")
	}

	runner := NewJobRunner(fanout)

	for i, host := range hostList {
		ch := make(chan string)
		runner.RunConfigureAuto(host, ch)
		if (i+1)%fanout == 0 {
			time.Sleep(time.Duration(delay) * time.Second)
		}
		output := strings.Split(<-ch, "|")

		if len(output) < 3 {
			return ToastError(f, nil, "Failed to run auto config job, index out of range")
		}
		h.writeEvent(output[0], f, fmt.Sprintf("%s: %s", output[1], output[2]))
	}
	runner.Wait()

	return ToastSuccess(f, "Successfully sent Auto Configure to node(s)", ``)
}
func (h *Handler) bmcConfigureImport(f *fiber.Ctx) error {
	delay := viper.GetInt("bmc.delay")
	fanout := viper.GetInt("bmc.fanout")
	file := f.FormValue("File")
	if file == "" {
		return ToastError(f, nil, "No file specified")
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

	runner := NewJobRunner(fanout)

	for i, host := range hostList {
		ch := make(chan string)
		runner.RunConfigureImport(host, file, ch)
		if (i+1)%fanout == 0 {
			time.Sleep(time.Duration(delay) * time.Second)
		}
		output := strings.Split(<-ch, "|")

		if len(output) < 3 {
			return ToastError(f, nil, "Failed to run import job, index out of range")
		}
		h.writeEvent(output[0], f, fmt.Sprintf("%s: %s", output[1], output[2]))
	}
	runner.Wait()

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
		err := h.DB.UpdateUser(user, role)
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
			if hostTableForm.Interfaces[i].BMC == "true" && (hostNameArr[0] == "cpn" || hostNameArr[0] == "srv") {
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
			ID:         ksuid.New(),
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
	for _, host := range newHosts {
		h.writeEvent("info", f, fmt.Sprintf("New host: %s added with BulkAdd", host.Name))
	}

	return ToastSuccess(f, "Successfully added host(s)", `, "closeModal": "", "refresh": ""`)
}
func (h *Handler) eventSSE(f *fiber.Ctx) error {
	f.Set("Content-Type", "text/event-stream")
	f.Set("Cache-Control", "no-cache")
	f.Set("Connection", "keep-alive")
	f.Set("Transfer-Encoding", "chunked")

	sent := 0
	if len(h.Events) > 5 {
		h.Events = h.Events[1:]
	}
	tdClasses := "border border-neutral-300 p-1"
	f.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		for {
			if sent >= len(h.Events) {
				time.Sleep(1 * time.Second)
				continue
			}
			e := h.Events[sent]
			msg := fmt.Sprintf(`<tr><td class="%s">%s</td><td class="%s">%s</td><td class="%s">%s</td><td class="%s">%s</td></tr>`, tdClasses, e.Time, tdClasses, e.User, tdClasses, e.Severity, tdClasses, e.Message)
			fmt.Fprintf(w, "data: %s\n\n", msg)

			err := w.Flush()
			if err != nil {
				log.Debugf("Error while flushing /events: %v. Closing http connection.\n", err)

				break
			}

			sent++
			time.Sleep(50 * time.Millisecond)
		}
	}))

	return nil
}
