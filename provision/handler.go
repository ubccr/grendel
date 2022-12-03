// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package provision

import (
	"bytes"
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
	"github.com/ubccr/grendel/model"
)

type Handler struct {
	DB               model.DataStore
	DefaultImageName string
}

func NewHandler(db model.DataStore, defaultImageName string) (*Handler, error) {
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
	if name == "" {
		return h.DB.LoadBootImage(h.DefaultImageName)
	}

	return h.DB.LoadBootImage(name)
}

func (h *Handler) SetupRoutes(e *echo.Echo) {
	e.GET("/", h.Index).Name = "index"

	boot := e.Group("/boot/:token/")
	boot.Use(TokenRequired)
	boot.POST("complete", h.Complete)
	boot.GET("ipxe", h.Ipxe)
	boot.GET("kickstart", h.Kickstart)
	boot.GET("file/kernel*", h.File)
	boot.GET("file/liveimg", h.File)
	boot.GET("file/rootfs", h.File)
	boot.GET("file/initrd-*", h.File)
	boot.GET("cloud-init/user-data", h.UserData)
	boot.GET("cloud-init/meta-data", h.MetaData)
	boot.GET("cloud-init/vendor-data", h.VendorData)
	boot.GET("pxe-config.ign", h.Ignition)
	boot.GET("provision/:name", h.ProvisionTemplate)
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
	endpoints := model.NewEndpoints(serverHost, token)

	data := map[string]interface{}{
		"token":     c.Param("token"),
		"endpoints": endpoints,
		"bootimage": bootImage,
		"nic":       nic,
		"host":      host,
		"rootpw":    viper.GetString("provision.root_password"),
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

	tmplName := "kickstart.tmpl"
	if bootImage.ProvisionTemplate != "" {
		tmplName = bootImage.ProvisionTemplate
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
			"id":   host.ID,
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

	tmplName := "user-data.tmpl"
	if bootImage.UserData != "" {
		tmplName = bootImage.UserData
	}

	log.Infof("Sending cloud-init user-data to host %s", host.Name)
	return c.Render(http.StatusOK, tmplName, data)
}

func (h *Handler) MetaData(c echo.Context) error {
	_, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	log.Infof("Sending cloud-init meta-data to host %s", host.Name)
	return c.Render(http.StatusOK, "meta-data.tmpl", data)
}

func (h *Handler) VendorData(c echo.Context) error {
	return c.String(http.StatusOK, "")
}

func (h *Handler) Ignition(c echo.Context) error {
	bootImage, host, _, data, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	tmplName := "butane.tmpl"
	if bootImage.Butane != "" {
		tmplName = bootImage.Butane
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
