package frontend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/segmentio/ksuid"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/firmware"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

func (h *Handler) LoginUser(c echo.Context) error {
	user := c.FormValue("username")
	pass := c.FormValue("password")

	val, err := h.DB.VerifyUser(user, pass)
	if err != nil || !val {
		msg := "Internal server error"
		if err.Error() == "crypto/bcrypt: hashedPassword is not the hash of the given password" {
			msg = ToastHtml("Failed to login: Password is incorrect", "error")
		} else if err.Error() == "not found" {
			msg = ToastHtml("Failed to login: Username not found.", "error")
		}

		return ToastError(c, err, msg)
	}
	if val {
		tCookie := new(http.Cookie)
		uCookie := new(http.Cookie)

		t, e, err := Sign(user, "user")
		if err != nil {
			return ToastError(c, err, "Internal server error")
		}
		tCookie.Name = "Authorization"
		tCookie.Value = fmt.Sprintf("Bearer %s", t)
		tCookie.HttpOnly = true
		tCookie.Secure = true
		tCookie.Expires = e
		tCookie.Path = "/"
		c.SetCookie(tCookie)

		uCookie.Name = "User"
		uCookie.Value = user
		uCookie.Expires = e
		uCookie.Path = "/"
		c.SetCookie(uCookie)

		c.Response().Header().Add("HX-Redirect", "/")
	}

	return ToastSuccess(c, "Successfully logged in")
}

func (h *Handler) LogoutUser(c echo.Context) error {
	t := new(http.Cookie)
	u := new(http.Cookie)

	t.Name = "Authorization"
	t.Value = ""
	t.Expires = time.Now()
	t.Path = "/"
	c.SetCookie(t)

	u.Name = "User"
	u.Value = ""
	u.Expires = time.Now()
	u.Path = "/"
	c.SetCookie(u)

	c.Response().Header().Add("HX-Redirect", "/")

	return ToastSuccess(c, "Successfully logged out")
}

func (h *Handler) RegisterUser(c echo.Context) error {
	u := c.FormValue("username")
	p := c.FormValue("password")

	err := h.DB.StoreUser(u, p)
	if err != nil {
		return ToastError(c, err, "Failed to register user")
	}

	return ToastSuccess(c, "Successfully registered user")
}

type FormData struct {
	ID         string
	Name       string
	Provision  string
	Firmware   string
	BootImage  string
	Tags       string
	Interfaces string
}

func (h *Handler) EditHost(c echo.Context) error {
	formHost := new(FormData)

	if err := c.Bind(formHost); err != nil {
		return ToastError(c, err, "Failed to bind type to request body")
	}
	if err := c.Validate(formHost); err != nil {
		return ToastError(c, err, "Failed to bind type to request body")
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

	err := h.DB.StoreHost(&newHost)
	if err != nil {
		return ToastError(c, err, "Failed to update host")
	}

	return ToastSuccess(c, "Successfully updated host")
}

// func (h *Handler) Provision(c echo.Context) error {
// 	reqHost, _ := nodeset.NewNodeSet(c.Param("nodeset"))
// 	host, _ := h.DB.FindHosts(reqHost)
// 	h.DB.ProvisionHosts(reqHost, !host[0].Provision)

// 	return c.HTML(http.StatusOK, fmt.Sprintf("%t", !host[0].Provision))
// }


type RebootData struct {
	Host string
}

func (h *Handler) RebootHost(c echo.Context) error {
	delay := viper.GetInt("bmc.delay")
	fanout := viper.GetInt("bmc.fanout")

	hosts := c.FormValue("hosts")
	ns, err := nodeset.NewNodeSet(hosts)
	if err != nil {
		return ToastError(c, err, "Failed to parse node set")
	}

	hostList, err := h.DB.FindHosts(ns)
	if err != nil {
		return ToastError(c, err, "Failed to find nodes")
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
	return ToastSuccess(c, "Successfully Rebooted node(s)")
}

func (h *Handler) BmcConfigure(c echo.Context) error {
	delay := viper.GetInt("bmc.delay")
	fanout := viper.GetInt("bmc.fanout")

	hosts := c.FormValue("hosts")
	ns, err := nodeset.NewNodeSet(hosts)
 	if err != nil {
	 return ToastError(c, err, "Failed to parse node set")
	}

	hostList, _ := h.DB.FindHosts(ns)
	if err != nil {
		return ToastError(c, err, "Failed to find nodes")
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
	return ToastSuccess(c, "Successfully sent Auto Configure node(s)")
}
