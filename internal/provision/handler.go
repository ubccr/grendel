// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/pkg/model"
)

type Handler struct {
	DB               store.Store
	DefaultImageName string
	netBoxClient     *http.Client
}

func init() {
	viper.SetDefault("provision.enable_prometheus_sd", false)
	viper.SetDefault("provision.prometheus_sd_refresh_interval", "3600")
}

func newNetBoxClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func NewHandler(db store.Store, defaultImageName string) (*Handler, error) {
	h := &Handler{
		DB:               db,
		DefaultImageName: defaultImageName,
		netBoxClient:     newNetBoxClient(),
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
	if name == "" {
		return h.DB.LoadBootImage(h.DefaultImageName)
	}

	return h.DB.LoadBootImage(name)
}

func (h *Handler) SetupRoutes(e *echo.Echo) {
	e.GET("/", h.Index).Name = "index"
	e.GET("/onie-installer*", h.Onie).Name = "onie"
	e.GET("/onie-updater*", h.Onie).Name = "onie"
	if viper.GetBool("provision.enable_prometheus_sd") {
		e.GET("/service-discovery/:tag/:port", h.ServiceDiscovery).Name = "sd"
		e.GET("/pdu-service-discovery/:tag/:port", h.PDUServiceDiscovery).Name = "psd"
	}

	boot := e.Group("/boot/:token/")
	boot.Use(TokenRequired)
	boot.POST("complete", h.Complete)
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
	if viper.IsSet("provision.netbox_token") && viper.IsSet("provision.netbox_url") {
		boot.GET("netbox/render-config", h.NetBoxRenderConfig)
	}
}

func (h *Handler) Index(c echo.Context) error {
	resp := map[string]interface{}{
		"status": "up",
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) verifyClaims(c echo.Context) (*model.BootImage, *model.Host, *model.NetInterface, map[string]interface{}, error) {
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

	token := c.Param("token")
	serverHost := c.Request().Host
	endpoints := NewEndpoints(serverHost, token)

	log.WithFields(logrus.Fields{
		"host":    host.Name,
		"headers": c.Request().Header,
	}).Debug("HTTP request headers")

	data := map[string]interface{}{
		"token":           c.Param("token"),
		"endpoints":       endpoints,
		"bootimage":       bootImage,
		"nic":             nic,
		"host":            host,
		"headers":         c.Request().Header,
		"rootpw":          viper.GetString("provision.root_password"),
		"adminSSHPubKeys": viper.GetStringSlice("admin_ssh_pubkeys"),
	}

	return bootImage, host, nic, data, nil
}

func (h *Handler) Ipxe(c echo.Context) error {
	bootImage, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	log.Infof("Sending iPXE script to boot host %s with image %s", host.Name, bootImage.Name)

	commandLine := bootImage.CommandLine

	if commandLine != "" {
		cmdTmpl, err := template.New("cmd").Parse(commandLine)
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		err = cmdTmpl.Execute(&buf, data)
		if err != nil {
			return err
		}
		commandLine = buf.String()
	}

	data["commandLine"] = commandLine

	return c.Render(http.StatusOK, "ipxe.tmpl", data)
}

func (h *Handler) File(c echo.Context) error {
	bootImage, host, _, _, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	_, fileType := path.Split(c.Request().URL.Path)

	log.Infof("Got request for file %q from host %s %s", fileType, host.Name, c.RealIP())

	switch {
	case fileType == "kernel":
		return c.File(bootImage.KernelPath)
	case fileType == "kernel.sig":
		return c.File(bootImage.KernelPath + ".sig")

	case fileType == "liveimg":
		return c.File(bootImage.LiveImage)

	case strings.HasPrefix(fileType, "initrd-"):
		initrdBaseName := strings.TrimSuffix(fileType, ".sig")
		i, err := strconv.Atoi(initrdBaseName[7:])
		if err != nil || i < 0 || i >= len(bootImage.InitrdPaths) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("no initrd with ID %q", i))
		}
		initrd := bootImage.InitrdPaths[i]
		if strings.HasSuffix(fileType, ".sig") {
			initrd += ".sig"
		}
		return c.File(initrd)
	}

	return echo.NewHTTPError(http.StatusNotFound, "")
}

func (h *Handler) serveBlob(c echo.Context, name string, data []byte) error {
	http.ServeContent(c.Response(), c.Request(), name, time.Time{}, bytes.NewReader(data))
	return nil
}

func (h *Handler) Kickstart(c echo.Context) error {
	bootImage, _, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	tmplName, ok := bootImage.ProvisionTemplates["kickstart"]
	if !ok {
		tmplName = "kickstart.tmpl"
	}

	return c.Render(http.StatusOK, tmplName, data)
}

func (h *Handler) Complete(c echo.Context) error {
	_, host, _, _, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	log.Infof("Unprovisioning host %s", host.Name)

	host.Provision = false

	err = h.DB.StoreHost(host)
	if err != nil {
		log.WithFields(logrus.Fields{
			"uid":  host.UID,
			"name": host.Name,
		}).Error("failed to unprovision host")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to unprovision host").SetInternal(err)
	}

	resp := map[string]interface{}{
		"status": "ok",
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) UserData(c echo.Context) error {
	bootImage, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	tmplName, ok := bootImage.ProvisionTemplates["user_data"]
	if !ok {
		tmplName = "user-data.tmpl"
	}

	log.Infof("Sending cloud-init user-data to host %s", host.Name)
	c.Response().Header().Set(echo.HeaderContentType, "application/yaml; charset=utf-8")
	return c.Render(http.StatusOK, tmplName, data)
}

func (h *Handler) MetaData(c echo.Context) error {
	_, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	log.Infof("Sending cloud-init meta-data to host %s", host.Name)
	c.Response().Header().Set(echo.HeaderContentType, "application/yaml; charset=utf-8")
	return c.Render(http.StatusOK, "meta-data.tmpl", data)
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

	tmplName, ok := bootImage.ProvisionTemplates["butane"]
	if !ok {
		tmplName = "butane.tmpl"
	}

	log.Infof("Sending ignition config to host %s", host.Name)
	renderer := c.Echo().Renderer.(*TemplateRenderer)
	return renderer.RenderIgnition(http.StatusOK, tmplName, data, c)
}

func (h *Handler) ProvisionTemplate(c echo.Context) error {
	bootImage, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	if bootImage.ProvisionTemplates == nil {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	tmplName, ok := bootImage.ProvisionTemplates[c.Param("name")]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	log.Infof("Sending provision template %s to host %s", c.Param("name"), host.Name)
	return c.Render(http.StatusOK, tmplName, data)
}

func (h *Handler) BmcTemplate(c echo.Context) error {
	_, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	tmplName := c.Param("name")
	if tmplName == "" {
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	log.Infof("Sending bmc template %s to host %s", c.Param("name"), host.Name)
	return c.Render(http.StatusOK, tmplName, data)
}

func (h *Handler) Proxmox(c echo.Context) error {
	bootImage, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	tmplName, ok := bootImage.ProvisionTemplates["proxmox"]
	if !ok {
		tmplName = "proxmox.tmpl"
	}

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
		return c.File(bootImage.KernelPath)
	}

	return echo.NewHTTPError(http.StatusBadRequest, "Invalid ONIE operation")
}
