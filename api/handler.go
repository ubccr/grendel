package api

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/model"
)

type Handler struct {
	BootSpec *model.BootSpec
}

func NewHandler(b *model.BootSpec) (*Handler, error) {
	return &Handler{BootSpec: b}, nil
}

func (h *Handler) SetupRoutes(e *echo.Echo) {
	e.GET("/", h.Index).Name = "index"
	e.GET("/_/ipxe", h.Ipxe).Name = "ipxe"
	e.GET("/_/file", h.File).Name = "file"
}

func (h *Handler) Index(c echo.Context) error {
	resp := map[string]interface{}{
		"status": "up",
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) Ipxe(c echo.Context) error {
	macStr := c.QueryParam("mac")
	if macStr == "" {
		log.WithFields(log.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
		}).Error("HTTP bad request missing MAC address")
		return echo.NewHTTPError(http.StatusBadRequest, "missing MAC address parameter")
	}
	mac, err := net.ParseMAC(macStr)
	if err != nil {
		log.WithFields(log.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
			"mac": macStr,
			"err": err,
		}).Error("HTTP bad request invalid mac")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid MAC address")
	}

	script, err := h.ipxeScript(mac, c.Request().Host, c.Scheme())
	log.WithFields(log.Fields{
		"mac": mac,
	}).Debug("Construct ipxe script")
	if err != nil {
		log.WithFields(log.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
			"mac": mac,
			"err": err,
		}).Error("Failed to assemble ipxe script")
		return echo.NewHTTPError(http.StatusInternalServerError, "couldn't get a boot script")
	}

	log.WithFields(log.Fields{
		"mac": mac,
		"ip":  c.RealIP(),
	}).Info("Sending ipxe boot script")
	fmt.Printf("\n\n%s\n\n", script)
	return c.String(http.StatusOK, string(script))
}

func (h *Handler) ipxeScript(mac net.HardwareAddr, serverHost, scheme string) ([]byte, error) {
	if h.BootSpec.Kernel == "" {
		return nil, errors.New("spec is missing Kernel")
	}

	urlTemplate := fmt.Sprintf("%s://%s/_/file?name=%%s&type=%%s&mac=%%s", scheme, serverHost)
	var b bytes.Buffer
	b.WriteString("#!ipxe\n")
	u := fmt.Sprintf(urlTemplate, url.QueryEscape(string(h.BootSpec.Kernel)), "kernel", url.QueryEscape(mac.String()))
	fmt.Fprintf(&b, "kernel --name kernel %s\n", u)
	for i, initrd := range h.BootSpec.Initrd {
		u = fmt.Sprintf(urlTemplate, url.QueryEscape(string(initrd)), "initrd", url.QueryEscape(mac.String()))
		fmt.Fprintf(&b, "initrd --name initrd%d %s\n", i, u)
	}

	b.WriteString("boot kernel ")
	for i := range h.BootSpec.Initrd {
		fmt.Fprintf(&b, "initrd=initrd%d ", i)
	}

	b.WriteString(h.BootSpec.Cmdline)
	b.WriteByte('\n')

	return b.Bytes(), nil
}

func (h *Handler) File(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		log.WithFields(log.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
		}).Error("HTTP bad request missing name")
		return echo.NewHTTPError(http.StatusBadRequest, "missing name")
	}

	log.Infof("Sending file %q to %s", name, c.RealIP())

	return c.File(name)
}
