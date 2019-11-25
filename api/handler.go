package api

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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
	if h.BootSpec.Kernel == nil {
		return nil, errors.New("spec is missing Kernel")
	}

	urlTemplate := fmt.Sprintf("%s://%s/_/file?type=%%s&mac=%%s", scheme, serverHost)
	var b bytes.Buffer
	b.WriteString("#!ipxe\n")
	u := fmt.Sprintf(urlTemplate, "kernel", url.QueryEscape(mac.String()))
	fmt.Fprintf(&b, "kernel --name kernel %s\n", u)
	for i := range h.BootSpec.Initrd {
		u = fmt.Sprintf(urlTemplate, fmt.Sprintf("initrd-%d", i), url.QueryEscape(mac.String()))
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
	path := c.QueryParam("type")
	if path == "" {
		log.WithFields(log.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
		}).Error("HTTP bad request missing type")
		return echo.NewHTTPError(http.StatusBadRequest, "missing type")
	}

	log.Infof("Got request for file %q to %s", path, c.RealIP())

	switch {
	case path == "kernel":
		return h.serveBlob(c, path, h.BootSpec.Kernel)

	case strings.HasPrefix(path, "initrd-"):
		i, err := strconv.Atoi(path[7:])
		if err != nil || i < 0 || i >= len(h.BootSpec.Initrd) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("no initrd with ID %q", i))
		}
		return h.serveBlob(c, path, h.BootSpec.Initrd[i])
	}

	return echo.NewHTTPError(http.StatusNotFound, "")
}

func (h *Handler) serveBlob(c echo.Context, name string, data []byte) error {
	http.ServeContent(c.Response(), c.Request(), name, time.Time{}, bytes.NewReader(data))
	return nil
}
