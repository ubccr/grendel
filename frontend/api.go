package frontend

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/segmentio/ksuid"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

func (h *Handler) LoginUser(f *fiber.Ctx) error {
	user := f.FormValue("username")
	pass := f.FormValue("password")

	val, err := h.DB.VerifyUser(user, pass)
	if err != nil || !val {
		msg := "Internal server error"
		if err.Error() == "crypto/bcrypt: hashedPassword is not the hash of the given password" {
			msg = "Failed to login: Password is incorrect"
		} else if err.Error() == "not found" {
			msg = "Failed to login: Username not found."
		}

		return ToastError(f, err, msg)
	}
	if val {
		tCookie := new(fiber.Cookie)
		uCookie := new(fiber.Cookie)

		t, e, err := Sign(user, "user")
		if err != nil {
			return ToastError(f, err, "Internal server error")
		}
		tCookie.Name = "Authorization"
		tCookie.Value = t
		tCookie.HTTPOnly = true
		tCookie.Secure = true
		tCookie.Expires = e
		tCookie.Path = "/"
		f.Cookie(tCookie)

		uCookie.Name = "User"
		uCookie.Value = user
		uCookie.Expires = e
		uCookie.Path = "/"
		f.Cookie(uCookie)

		f.Response().Header.Add("HX-Redirect", "/")
	}

	return ToastSuccess(f, "Successfully logged in")
}

func (h *Handler) LogoutUser(f *fiber.Ctx) error {
	t := new(fiber.Cookie)
	u := new(fiber.Cookie)

	t.Name = "Authorization"
	t.Value = ""
	t.Expires = time.Now()
	t.Path = "/"
	f.Cookie(t)

	u.Name = "User"
	u.Value = ""
	u.Expires = time.Now()
	u.Path = "/"
	f.Cookie(u)

	f.Response().Header.Add("HX-Redirect", "/")

	return ToastSuccess(f, "Successfully logged out")
}

func (h *Handler) RegisterUser(f *fiber.Ctx) error {
	u := f.FormValue("username")
	p := f.FormValue("password")

	err := h.DB.StoreUser(u, p)
	if err != nil {
		return ToastError(f, err, "Failed to register user")
	}

	return ToastSuccess(f, "Successfully registered user")
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

func (h *Handler) EditHost(f *fiber.Ctx) error {
	formHost := new(FormData)

	err := f.BodyParser(formHost)
	if err != nil {
		return ToastError(f, err, "Failed to bind type to request body")
	}

	id, _ := ksuid.Parse(formHost.ID)

	provision := false
	if formHost.Provision == "on" {
		provision = true
	}
	var ifaces []*model.NetInterface
	json.Unmarshal([]byte(formHost.Interfaces), &ifaces)

	newHost := model.Host{
		ID:         id,
		Name:       formHost.Name,
		Provision:  provision,
		Firmware:   firmware.NewFromString(formHost.Firmware),
		BootImage:  formHost.BootImage,
		Tags:       strings.Split(formHost.Tags, ","),
		Interfaces: ifaces,
	}

	err = h.DB.StoreHost(&newHost)
	if err != nil {
		return ToastError(f, err, "Failed to update host")
	}

	return ToastSuccess(f, "Successfully updated host")
}

func (h *Handler) DeleteHost(f *fiber.Ctx) error {
	hosts := f.FormValue("hosts")
	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(f, err, "Failed to parse node set")
	}

	h.DB.DeleteHosts(ns)

	f.Response().Header.Add("HX-Refresh", "true")
	return ToastSuccess(f, "Successfully deleted host(s)")
}

type RebootData struct {
	Host string
}

func (h *Handler) RebootHost(f *fiber.Ctx) error {
	delay := viper.GetInt("bmc.delay")
	fanout := viper.GetInt("bmc.fanout")

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

	for i, host := range hostList {
		runner.RunReboot(host)
		if (i+1)%fanout == 0 {
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}
	runner.Wait()

	// TODO: add channel to get status of each job
	return ToastSuccess(f, "Successfully Rebooted node(s)")
}

func (h *Handler) BmcConfigure(f *fiber.Ctx) error {
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
		runner.RunConfigure(host)
		if (i+1)%fanout == 0 {
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}
	runner.Wait()

	// TODO: add channel to get status of each job
	return ToastSuccess(f, "Successfully sent Auto Configure node(s)")
}

func (h *Handler) HostAdd(f *fiber.Ctx) error {
	fw := f.FormValue("Firmware")
	p := f.FormValue("Provision")
	bootImage := f.FormValue("BootImage")
	tags := f.FormValue("Tags")

	provision, _ := strconv.ParseBool(p)
	hostList := strings.Split(f.FormValue("HostList"), ",")

	for _, v := range hostList {
		mgmtMac, err := net.ParseMAC(f.FormValue(fmt.Sprintf("%s:MgmtMac", v)))
		if err != nil {
			return ToastError(f, err, "Failed to parse MAC address")
		}
		coreMac, err := net.ParseMAC(f.FormValue(fmt.Sprintf("%s:CoreMac", v)))
		if err != nil {
			return ToastError(f, err, "Failed to parse MAC address")
		}

		ifaces := make([]*model.NetInterface, 0)
		ifaces = append(ifaces, &model.NetInterface{
			Name: "",
			FQDN: "",
			MAC:  mgmtMac,
			BMC:  true,
			VLAN: "",
			MTU:  1500,
		})
		ifaces = append(ifaces, &model.NetInterface{
			Name: "",
			FQDN: fmt.Sprintf("%s.core.ccr.buffalo.edu", v),
			MAC:  coreMac,
			BMC:  false,
			VLAN: "",
			MTU:  9000,
		})
		newHost := model.Host{
			ID:        ksuid.New(),
			Name:      v,
			Provision: provision,
			BootImage: bootImage,
			Firmware:  firmware.NewFromString(fw),
			Tags:      strings.Split(tags, ","),
		}
		err = h.DB.StoreHost(&newHost)
		if err != nil {
			return ToastError(f, err, "Failed to add host(s)")
		}

	}

	f.Response().Header.Add("HX-Refresh", "true")
	return ToastSuccess(f, "Successfully added host(s)")
}

func (h *Handler) SwitchMac(f *fiber.Ctx) error {
	rack := f.FormValue("Rack")
	hosts := strings.Split(f.FormValue("HostList"), ",")
	mgmtMacList, err := GetMacAddress(h, f, rack, "Mgmt", hosts)
	if err != nil {
		return ToastError(f, err, "Failed to get MAC address")
	}
	coreMacList, err := GetMacAddress(h, f, rack, "Core", hosts)
	if err != nil {
		return ToastError(f, err, "Failed to get MAC address")
	}

	type hostStruct struct {
		Host     string
		MgmtPort string
		CorePort string
		MgmtMac  string
		CoreMac  string
	}
	hostList := make([]hostStruct, len(hosts))

	for i, host := range hosts {
		hostList[i] = hostStruct{
			Host:     host,
			MgmtPort: f.FormValue(fmt.Sprintf("%s:Mgmt", host), ""),
			CorePort: f.FormValue(fmt.Sprintf("%s:Core", host), ""),
			MgmtMac:  mgmtMacList[host],
			CoreMac:  coreMacList[host],
		}
	}

	return f.Render("fragments/hostAddModalList", fiber.Map{
		"Hosts":    hostList,
		"HostList": strings.Join(hosts, ","),
	}, "")
}
