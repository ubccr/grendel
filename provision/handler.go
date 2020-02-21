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
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	boot := e.Group("/boot/")

	config := middleware.JWTConfig{
		Claims:      &model.BootClaims{},
		ContextKey:  ContextKeyJWT,
		SigningKey:  []byte(viper.GetString("provision.secret")),
		TokenLookup: "query:token",
	}
	boot.Use(middleware.JWTWithConfig(config))
	boot.GET("ipxe", h.Ipxe)
	boot.GET("kickstart", h.Kickstart)
	boot.GET("file/kernel*", h.File)
	boot.GET("file/liveimg", h.File)
	boot.GET("file/rootfs", h.File)
	boot.GET("file/initrd-*", h.File)
}

func (h *Handler) Index(c echo.Context) error {
	resp := map[string]interface{}{
		"status": "up",
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) Ipxe(c echo.Context) error {
	bootToken := c.Get(ContextKeyJWT).(*jwt.Token)
	claims := bootToken.Claims.(*model.BootClaims)

	log.Debugf("iPXE Got valid boot claims: %v", claims)

	host, err := h.DB.LoadHostFromID(claims.ID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("iPXE failed to find host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid host")
	}

	mac, err := net.ParseMAC(claims.MAC)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("iPXE got invalid mac address")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid mac address")
	}

	nic := host.Interface(mac)
	if nic == nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("iPXE got invalid boot interface for host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid boot interface")
	}

	bootImage, err := h.LoadBootImageWithDefault(host.BootImage)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("iPXE failed to find boot image for host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid boot image")
	}

	log.Infof("Sending iPXE script to boot host %s with image %s", host.Name, bootImage.Name)
	baseURI := fmt.Sprintf("%s://%s", c.Scheme(), c.Request().Host)
	commandLine := bootImage.CommandLine

	if host.Kickstart {
		commandLine += fmt.Sprintf(" ks=%s/boot/kickstart?token=%s network ksdevice=bootif ks.device=bootif inst.stage2=%s/repo/%s ", baseURI, c.QueryParam("token"), baseURI, bootImage.InstallRepo)
		if viper.GetBool("provision.noverifyssl") {
			commandLine += " rd.noverifyssl noverifyssl inst.noverifyssl"
		}
	} else if len(bootImage.LiveImage) > 0 && !strings.Contains(bootImage.CommandLine, "live") {
		commandLine += fmt.Sprintf(" root=live:%s/boot/file/liveimg?token=%s", baseURI, c.QueryParam("token"))
		if viper.GetBool("provision.noverifyssl") {
			commandLine += " rd.noverifyssl"
		}
	}

	data := map[string]interface{}{
		"token":       c.QueryParam("token"),
		"bootimage":   bootImage,
		"commandLine": commandLine,
		"nic":         nic,
		"host":        host,
		"baseuri":     baseURI,
	}

	return c.Render(http.StatusOK, "ipxe.tmpl", data)
}

func (h *Handler) File(c echo.Context) error {
	bootToken := c.Get(ContextKeyJWT).(*jwt.Token)
	claims := bootToken.Claims.(*model.BootClaims)

	log.Debugf("File handler Got valid boot claims: %v", claims)

	macStr := claims.MAC
	if macStr == "" {
		log.WithFields(logrus.Fields{
			"ip": c.RealIP(),
		}).Error("Bad request missing MAC address")
		return echo.NewHTTPError(http.StatusBadRequest, "missing MAC address parameter")
	}

	host, err := h.DB.LoadHostFromID(claims.ID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("Bad request failed to find host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid host")
	}

	bootImage, err := h.LoadBootImageWithDefault(host.BootImage)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("Bad request failed to find boot image for host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid boot image")
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
	bootToken := c.Get(ContextKeyJWT).(*jwt.Token)
	claims := bootToken.Claims.(*model.BootClaims)

	log.Infof("Kickstart got valid boot claims: %v", claims)

	host, err := h.DB.LoadHostFromID(claims.ID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("Kickstart failed to find host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid host")
	}

	bootImage, err := h.LoadBootImageWithDefault(host.BootImage)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("Kickstart failed to find boot image for host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid boot image")
	}

	mac, err := net.ParseMAC(claims.MAC)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("Kickstart got invalid mac address")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid mac address")
	}

	nic := host.Interface(mac)
	if nic == nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("Kickstart got invalid boot interface for host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid boot interface")
	}

	baseURI := fmt.Sprintf("%s://%s", c.Scheme(), c.Request().Host)

	data := map[string]interface{}{
		"token":     c.QueryParam("token"),
		"bootimage": bootImage,
		"baseuri":   baseURI,
		"host":      host,
		"nic":       nic,
		"rootpw":    viper.GetString("provision.root_password"),
	}

	return c.Render(http.StatusOK, "kickstart.tmpl", data)
}
