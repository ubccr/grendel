// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"bytes"
	"cmp"
	"errors"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/internal/render"
	"github.com/ubccr/grendel/internal/settings"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/netbox"
)

type Handler struct {
	DB               store.Store
	DefaultImageName string
	Templates        *render.Set
}

func NewHandler(db store.Store, defaultImageName string) (*Handler, error) {
	h := &Handler{
		DB:               db,
		DefaultImageName: defaultImageName,
	}

	if defaultImageName != "" {
		_, err := h.DB.LoadBootImage(defaultImageName)
		if err != nil {
			return nil, err

		}
	}

	return h, nil
}

func (h *Handler) LoadBootImageWithDefault(name string) (*model.BootImage, error) {
	// the setting is read each time so a new default applies without a restart, and the name the handler was built with stands in when it is unset
	defaultImage := cmp.Or(settings.ProvisionDefaultImage.Get(), h.DefaultImageName)
	if name == "" && defaultImage == "" {
		log.Warn("Cannot find boot image! Please either set a boot_image on the node or the provision.default_image setting")
	} else if name == "" {
		name = defaultImage
	}

	return h.DB.LoadBootImage(name)
}

func (h *Handler) SetupRoutes(e *echo.Echo) {
	e.GET("/", h.Index).Name = "index"
	e.GET("/onie-installer*", h.Onie).Name = "onie"
	e.GET("/onie-updater*", h.Onie).Name = "onie"
	sd := e.Group("", prometheusSDEnabled)
	sd.GET("/service-discovery/:tag/:port", h.ServiceDiscovery).Name = "sd"
	sd.GET("/pdu-service-discovery/:tag/:port", h.PDUServiceDiscovery).Name = "psd"

	e.GET("/beacon/:arch", h.Beacon).Name = "beacon"

	boot := e.Group("/boot/:token/")
	boot.Use(TokenRequired)
	boot.POST("complete", h.Complete)
	boot.POST("status", h.Status)
	boot.GET("ipxe", h.Ipxe)
	boot.GET("kickstart", h.Kickstart)
	boot.GET("file/kernel*", h.File)
	boot.HEAD("file/liveimg", h.File)
	boot.GET("file/liveimg", h.File)
	boot.GET("file/rootfs", h.File)
	boot.GET("file/initrd-*", h.File)
	boot.GET("cloud-init/user-data", h.UserData)
	boot.GET("cloud-init/meta-data", h.MetaData)
	boot.GET("cloud-init/vendor-data", h.VendorData)
	boot.GET("pxe-config.ign", h.Ignition)
	boot.GET("provision/:name", h.ProvisionTemplate)
	boot.GET("bmc/:name", h.BmcTemplate)
	boot.POST("proxmox", h.Proxmox)
	if settings.ProvisionNetboxToken.Get() != "" && settings.ProvisionNetboxURL.Get() != "" {
		boot.GET("netbox/render-config", h.NetBoxRenderConfig)
	}
}

// prometheusSDEnabled answers 404 while service discovery is turned off, which is checked per request so turning it on needs no restart
func prometheusSDEnabled(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !settings.ProvisionEnablePrometheusSD.Get() {
			return echo.ErrNotFound
		}
		return next(c)
	}
}

func (h *Handler) Index(c echo.Context) error {
	resp := map[string]any{
		"status": "up",
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) verifyClaims(c echo.Context) (*model.BootImage, *model.Host, *model.NetInterface, map[string]any, error) {
	claims := c.Get(ContextKeyToken).(*model.BootClaims)

	log.Debugf("Got valid boot claims: %v", claims)

	host, err := h.DB.LoadHostFromID(claims.ID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"host_id": claims.ID,
			"mac":     claims.MAC,
		}).Error("failed to find host")
		return nil, nil, nil, nil, echo.NewHTTPError(http.StatusBadRequest, "invalid host").SetInternal(err)
	}

	if !host.Provision {
		log.WithFields(logrus.Fields{
			"host_id": claims.ID,
			"mac":     claims.MAC,
		}).Error("host is not set to provision")
		return nil, nil, nil, nil, echo.NewHTTPError(http.StatusBadRequest, "host not set to provision")
	}

	mac, err := net.ParseMAC(claims.MAC)
	if err != nil {
		log.WithFields(logrus.Fields{
			"host_id": claims.ID,
			"mac":     claims.MAC,
		}).Error("got invalid mac address")
		return nil, nil, nil, nil, echo.NewHTTPError(http.StatusBadRequest, "invalid mac address").SetInternal(err)
	}

	nic := host.Interface(mac)
	if nic == nil {
		log.WithFields(logrus.Fields{
			"host_id": claims.ID,
			"mac":     claims.MAC,
		}).Error("got invalid boot interface for host")
		return nil, nil, nil, nil, echo.NewHTTPError(http.StatusBadRequest, "invalid boot interface").SetInternal(err)
	}

	bootImage, err := h.LoadBootImageWithDefault(host.BootImage)
	if err != nil {
		log.WithFields(logrus.Fields{
			"host_id": claims.ID,
			"mac":     claims.MAC,
		}).Error("failed to find boot image for host")
		return nil, nil, nil, nil, echo.NewHTTPError(http.StatusBadRequest, "invalid boot image").SetInternal(err)
	}

	// A manually set arch takes precedence over the one detected during the PXE boot
	if host.Arch == "" {
		host.Arch = claims.Arch
	}

	log.WithFields(logrus.Fields{
		"host":    host.Name,
		"headers": c.Request().Header,
	}).Debug("HTTP request headers")

	data := render.Data(render.Request{
		Host:    host,
		NIC:     nic,
		Image:   bootImage,
		Server:  c.Request().Host,
		Token:   c.Param("token"),
		Headers: c.Request().Header,
	})

	return bootImage, host, nic, data, nil
}

func (h *Handler) Ipxe(c echo.Context) error {
	bootImage, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	tmplName, _ := render.TemplateFor(bootImage, render.KindIPXE)

	if err := render.AddCommandLine(bootImage, data); err != nil {
		return err
	}

	log.Infof("iPXE script requested: script=%s host=%s boot_image=%s", tmplName, host.Name, bootImage.Name)
	if err := c.Render(http.StatusOK, tmplName, data); err != nil {
		return err
	}
	h.recordStage(host, model.ProvisionStateIPXE, bootImage)
	return nil
}

func (h *Handler) File(c echo.Context) error {
	bootImage, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	_, fileType := path.Split(c.Request().URL.Path)

	log.Infof("Got request for file %q from host %s %s", fileType, host.Name, c.RealIP())

	if fileType == "liveimg" {
		return c.File(bootImage.LiveImage)
	}

	// a signature is served from beside the file it signs
	name, isSig := strings.CutSuffix(fileType, ".sig")
	filePath, err := h.imagePath(name, host, bootImage, data)
	if err != nil {
		return err
	}
	if isSig {
		filePath += ".sig"
	}
	return c.File(filePath)
}

// imagePath is where the boot file the image serves as name is on disk for host
func (h *Handler) imagePath(name string, host *model.Host, bootImage *model.BootImage, data map[string]any) (string, error) {
	rendered, err := render.FilePath(bootImage, name, data)
	if errors.Is(err, render.ErrNoFile) {
		return "", echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	if err != nil {
		log.WithFields(logrus.Fields{
			"host":      host.Name,
			"bootimage": bootImage.Name,
			"file":      name,
			"err":       err,
		}).Error("failed to render boot image path template")
		return "", echo.NewHTTPError(http.StatusInternalServerError, "failed to render boot image path").SetInternal(err)
	}

	return rendered, nil
}

func (h *Handler) serveBlob(c echo.Context, name string, data []byte) error {
	http.ServeContent(c.Response(), c.Request(), name, time.Time{}, bytes.NewReader(data))
	return nil
}

func (h *Handler) Kickstart(c echo.Context) error {
	bootImage, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	tmplName, _ := render.TemplateFor(bootImage, render.KindKickstart)

	if err := c.Render(http.StatusOK, tmplName, data); err != nil {
		return err
	}
	h.recordStage(host, model.ProvisionStateConfig, bootImage)
	return nil
}

// Complete marks the node's provision completed and turns provisioning off. With ?unprovision=false it only records the status, for nodes that network boot every time
func (h *Handler) Complete(c echo.Context) error {
	_, host, _, _, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	unprovision := true
	if v := c.QueryParam("unprovision"); v != "" {
		unprovision, err = strconv.ParseBool(v)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "unprovision must be true or false")
		}
	}

	if unprovision {
		if err := h.unprovision(host); err != nil {
			return err
		}
	}
	h.recordCompleted(host)

	resp := map[string]any{
		"status": "ok",
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) UserData(c echo.Context) error {
	bootImage, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	tmplName, _ := render.TemplateFor(bootImage, render.KindUserData)

	log.Infof("Sending cloud-init user-data to host %s", host.Name)
	c.Response().Header().Set(echo.HeaderContentType, "application/yaml; charset=utf-8")
	if err := c.Render(http.StatusOK, tmplName, data); err != nil {
		return err
	}
	h.recordStage(host, model.ProvisionStateConfig, bootImage)
	return nil
}

func (h *Handler) MetaData(c echo.Context) error {
	_, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	log.Infof("Sending cloud-init meta-data to host %s", host.Name)
	c.Response().Header().Set(echo.HeaderContentType, "application/yaml; charset=utf-8")
	return c.Render(http.StatusOK, render.MetaDataTemplate, data)
}

func (h *Handler) VendorData(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/yaml; charset=utf-8")
	return c.String(http.StatusOK, "")
}

func (h *Handler) Ignition(c echo.Context) error {
	bootImage, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	tmplName, _ := render.TemplateFor(bootImage, render.KindButane)

	log.Infof("Sending ignition config to host %s", host.Name)
	ignition, err := h.Templates.Ignition(tmplName, data)
	if err != nil {
		if errors.Is(err, render.ErrButane) {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to translate butane config").SetInternal(err)
		}
		return err
	}
	if err := c.HTMLBlob(http.StatusOK, ignition); err != nil {
		return err
	}
	h.recordStage(host, model.ProvisionStateConfig, bootImage)
	return nil
}

func (h *Handler) ProvisionTemplate(c echo.Context) error {
	templateName := c.Param("name")

	bootImage, node, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	if bootImage.ProvisionTemplates == nil {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	template, ok := bootImage.ProvisionTemplates[templateName]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	log.WithFields(logrus.Fields{"template": templateName, "node": node.Name}).Info("Sending provision template")
	return c.Render(http.StatusOK, template, data)
}

func (h *Handler) BmcTemplate(c echo.Context) error {
	_, node, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	templateName := c.Param("name")
	if templateName == "" {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	log.WithFields(logrus.Fields{"template": templateName, "node": node.Name}).Info("Sending bmc template")
	return c.Render(http.StatusOK, templateName, data)
}

func (h *Handler) Proxmox(c echo.Context) error {
	bootImage, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	tmplName, _ := render.TemplateFor(bootImage, render.KindProxmox)

	log.Infof("Sending automated install answer file to host %s", host.Name)
	c.Response().Header().Set(echo.HeaderContentType, "application/yaml; charset=utf-8")
	return c.Render(http.StatusOK, tmplName, data)
}

func (h *Handler) Onie(c echo.Context) error {

	onie, err := NewOnieFromHeaders(c.Request().Header)
	if err != nil {
		log.WithFields(logrus.Fields{
			"msg": err,
		}).Error("Failed to parse ONIE headers")
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	host, err := h.DB.LoadHostFromMAC(onie.MAC.String())
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			log.Debugf("Ignoring unknown host mac address: %s", onie.MAC)
		} else {
			log.Errorf("ONIE handler failed to fetch host from database for mac %s: %s", onie.MAC, err)
		}
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	if !host.Provision {
		return echo.NewHTTPError(http.StatusNotFound, "ONIE install requested but host not set to provision")
	}

	bootImage, err := h.LoadBootImageWithDefault(host.BootImage)
	if err != nil {
		log.WithFields(logrus.Fields{
			"host": host.Name,
			"mac":  onie.MAC.String(),
		}).Error("failed to find boot image for host")
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	log.WithFields(logrus.Fields{
		"host":                host.Name,
		"mac":                 onie.MAC.String(),
		"bootimage":           bootImage.Name,
		"onieOp":              onie.Operation,
		"onieVendor":          onie.VendorID,
		"onieMachine":         onie.Machine,
		"onieRev":             onie.MachineRev,
		"onieArch":            onie.Arch,
		"onieUpdaterFilePath": onie.UpdaterFilePath(),
	}).Info("ONIE Request Data")

	switch onie.Operation {
	case OnieUpdate:
		return c.File(onie.UpdaterFilePath())
	case OnieInstall:
		data := map[string]any{
			"bootimage":  bootImage,
			"host":       host,
			"headers":    c.Request().Header,
			"extra_vars": settings.ProvisionExtraVars.Get(),
		}
		kernel, err := h.imagePath("kernel", host, bootImage, data)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "")
		}
		return c.File(kernel)
	}

	return echo.NewHTTPError(http.StatusBadRequest, "Invalid ONIE operation")
}

func (h *Handler) NetBoxRenderConfig(c echo.Context) error {
	_, node, _, _, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	res, err := render.NetBox.NetBoxFetchConfig(node.Name)
	if err != nil {
		if errors.Is(err, netbox.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render config").SetInternal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.WithFields(logrus.Fields{
			"name": node.Name,
			"code": res.StatusCode,
		}).Error("failed to fetch netbox render-config wrong http code")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render config").SetInternal(err)
	}

	return c.Stream(http.StatusOK, "text/plain", res.Body)
}
