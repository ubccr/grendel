package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.universe.tf/netboot/pixiecore"
)

type Handler struct {
	Booter pixiecore.Booter
}

func NewHandler(b pixiecore.Booter) (*Handler, error) {
	return &Handler{Booter: b}, nil
}

func (h *Handler) SetupRoutes(e *echo.Echo) {
	e.GET("/", h.Index).Name = "index"
	e.GET("/_/ipxe", h.Ipxe).Name = "ipxe"
	e.GET("/_/booting", h.Booting).Name = "booting"
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
		logrus.WithFields(logrus.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
		}).Error("HTTP bad request missing MAC address")
		return echo.NewHTTPError(http.StatusBadRequest, "missing MAC address parameter")
	}
	archStr := c.QueryParam("arch")
	if archStr == "" {
		logrus.WithFields(logrus.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
		}).Error("HTTP bad request missing architecture")
		return echo.NewHTTPError(http.StatusBadRequest, "missing architecture parameter")
	}

	mac, err := net.ParseMAC(macStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
			"mac": macStr,
			"err": err,
		}).Error("HTTP bad request invalid mac")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid MAC address")
	}

	i, err := strconv.Atoi(archStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url":  c.Request().URL,
			"ip":   c.RealIP(),
			"arch": archStr,
			"err":  err,
		}).Error("HTTP bad request invalid architecture")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid architecture")
	}
	arch := pixiecore.Architecture(i)
	switch arch {
	case pixiecore.ArchIA32, pixiecore.ArchX64:
	default:
		logrus.WithFields(logrus.Fields{
			"url":  c.Request().URL,
			"ip":   c.RealIP(),
			"arch": arch,
			"err":  err,
		}).Error("HTTP bad request unknown architecture")
		return echo.NewHTTPError(http.StatusBadRequest, "unknown architecture")
	}

	mach := pixiecore.Machine{
		MAC:  mac,
		Arch: arch,
	}
	start := time.Now()
	spec, err := h.Booter.BootSpec(mach)
	logrus.WithFields(logrus.Fields{
		"time": time.Since(start),
		"mac":  mac,
	}).Debug("Got bootspec")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
			"mac": mac,
			"err": err,
		}).Error("couldn't get bootspec")
		return echo.NewHTTPError(http.StatusInternalServerError, "couldn't get bootspec")
	}
	if spec == nil {
		// TODO: make ipxe abort netbooting so it can fall through to
		// other boot options - unsure if that's possible.
		logrus.WithFields(logrus.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
			"mac": mac,
		}).Info("you don't netboot")
		return echo.NewHTTPError(http.StatusNotFound, "you don't netboot")
	}
	start = time.Now()

	script, err := ipxeScript(mach, spec, c.Request().Host, c.Scheme())
	logrus.WithFields(logrus.Fields{
		"time": time.Since(start),
		"mac":  mac,
	}).Debug("Construct ipxe script")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
			"mac": mac,
			"err": err,
		}).Error("Failed to assemble ipxe script")
		return echo.NewHTTPError(http.StatusInternalServerError, "couldn't get a boot script")
	}

	logrus.WithFields(logrus.Fields{
		"mac": mac,
		"ip":  c.RealIP(),
	}).Info("Sending ipxe boot script")
	//s.machineEvent(mac, machineStateIpxeScript, "Sent iPXE boot script")
	fmt.Printf("\n\n%s\n\n", script)
	return c.String(http.StatusOK, string(script))
	//s.debug("HTTP", "Writing ipxe script to %s took %s", mac, time.Since(start))
	//s.debug("HTTP", "handleIpxe for %s took %s", mac, time.Since(overallStart))
}

func ipxeScript(mach pixiecore.Machine, spec *pixiecore.Spec, serverHost, scheme string) ([]byte, error) {
	if spec.IpxeScript != "" {
		return []byte(spec.IpxeScript), nil
	}

	if spec.Kernel == "" {
		return nil, errors.New("spec is missing Kernel")
	}

	urlTemplate := fmt.Sprintf("%s://%s/_/file?name=%%s&type=%%s&mac=%%s", scheme, serverHost)
	var b bytes.Buffer
	b.WriteString("#!ipxe\n")
	u := fmt.Sprintf(urlTemplate, url.QueryEscape(string(spec.Kernel)), "kernel", url.QueryEscape(mach.MAC.String()))
	fmt.Fprintf(&b, "kernel --name kernel %s\n", u)
	for i, initrd := range spec.Initrd {
		u = fmt.Sprintf(urlTemplate, url.QueryEscape(string(initrd)), "initrd", url.QueryEscape(mach.MAC.String()))
		fmt.Fprintf(&b, "initrd --name initrd%d %s\n", i, u)
	}

	//fmt.Fprintf(&b, "imgfetch --name ready %s://%s/_/booting?mac=%s ||\n", scheme, serverHost, url.QueryEscape(mach.MAC.String()))
	//b.WriteString("imgfree ready ||\n")

	b.WriteString("boot kernel ")
	for i := range spec.Initrd {
		fmt.Fprintf(&b, "initrd=initrd%d ", i)
	}

	f := func(id string) string {
		return fmt.Sprintf("%s://%s/_/file?name=%s", scheme, serverHost, url.QueryEscape(id))
	}
	cmdline, err := expandCmdline(spec.Cmdline, template.FuncMap{"ID": f})
	if err != nil {
		return nil, fmt.Errorf("expanding cmdline %q: %s", spec.Cmdline, err)
	}
	b.WriteString(cmdline)
	b.WriteByte('\n')

	return b.Bytes(), nil
}

func expandCmdline(tpl string, funcs template.FuncMap) (string, error) {
	tmpl, err := template.New("cmdline").Option("missingkey=error").Funcs(funcs).Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parsing cmdline %q: %s", tpl, err)
	}
	var out bytes.Buffer
	if err = tmpl.Execute(&out, nil); err != nil {
		return "", fmt.Errorf("expanding cmdline template %q: %s", tpl, err)
	}
	cmdline := strings.TrimSpace(out.String())
	if strings.Contains(cmdline, "\n") {
		return "", fmt.Errorf("cmdline %q contains a newline", cmdline)
	}
	return cmdline, nil
}

func (h *Handler) Booting(c echo.Context) error {
	// Return a no-op boot script, to satisfy iPXE. It won't get used,
	// the boot script deletes this image immediately after
	// downloading.
	macStr := c.QueryParam("mac")
	if macStr == "" {
		logrus.WithFields(logrus.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
		}).Error("HTTP bad request missing MAC address")
		return echo.NewHTTPError(http.StatusBadRequest, "missing MAC address parameter")
	}
	_, err := net.ParseMAC(macStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
			"mac": macStr,
			"err": err,
		}).Error("HTTP bad request invalid mac")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid MAC address")
	}
	return c.String(http.StatusOK, "# Booting")
	//s.machineEvent(mac, machineStateBooted, "Booting into OS")
}

func (h *Handler) File(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		logrus.WithFields(logrus.Fields{
			"url": c.Request().URL,
			"ip":  c.RealIP(),
		}).Error("HTTP bad request missing name")
		return echo.NewHTTPError(http.StatusBadRequest, "missing name")
	}

	f, sz, err := h.Booter.ReadBootFile(pixiecore.ID(name))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"url":  c.Request().URL,
			"ip":   c.RealIP(),
			"name": name,
		}).Error("HTTP failed getting file")
		return echo.NewHTTPError(http.StatusInternalServerError, "couldn't get file")
	}
	defer f.Close()
	if sz >= 0 {
		c.Response().Header().Set("Content-Length", strconv.FormatInt(sz, 10))
	} else {
		logrus.Warn("HTTP", "Unknown file size for %q, boot will be VERY slow (can your Booter provide file sizes?)", name)
	}
	if _, err = io.Copy(c.Response(), f); err != nil {
		logrus.WithFields(logrus.Fields{
			"url":  c.Request().URL,
			"ip":   c.RealIP(),
			"name": name,
		}).Error("HTTP failed writing file")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed writing file")
	}
	logrus.Infof("Sent file %q to %s", name, c.RealIP())

	return nil
}
