package frontend

import (
	"encoding/json"
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
			msg = ToastHtml("Failed to login: Password is incorrect", "error")
		} else if err.Error() == "not found" {
			msg = ToastHtml("Failed to login: Username not found.", "error")
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
