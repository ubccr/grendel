package api

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
	DB model.Datastore
}

func NewHandler(db model.Datastore) (*Handler, error) {
	h := &Handler{
		DB: db,
	}

	return h, nil
}

func (h *Handler) SetupRoutes(e *echo.Echo) {
	e.GET("/", h.Index).Name = "index"

	boot := e.Group("/_/")

	config := middleware.JWTConfig{
		Claims:      &model.BootClaims{},
		ContextKey:  ContextKeyJWT,
		SigningKey:  []byte(viper.GetString("secret")),
		TokenLookup: "query:token",
	}
	boot.Use(middleware.JWTWithConfig(config))
	boot.GET("ipxe", h.Ipxe)
	boot.GET("kickstart", h.Kickstart)
	boot.GET("file/kernel", h.File)
	boot.GET("file/liveimg", h.File)
	boot.GET("file/rootfs", h.File)
	boot.GET("file/initrd-*", h.File)

	v1 := e.Group("/v1/")
	v1.POST("host/add", h.HostAdd)
	v1.GET("host/list", h.HostList)
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

	log.Debugf("iPXE Got valid boot claims %s: %v", c.QueryParam("token"), claims)

	bootImage, err := h.DB.GetBootImage(claims.ID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("iPXE failed to find boot image for host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid boot image")
	}

	host, err := h.DB.GetHost(claims.ID)
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

	baseURI := fmt.Sprintf("%s://%s", c.Scheme(), c.Request().Host)
	commandLine := bootImage.CommandLine

	if host.Provision {
		commandLine += fmt.Sprintf(" ks=%s/_/kickstart?token=%s network ksdevice=bootif ks.device=bootif inst.stage2=%s ", baseURI, c.QueryParam("token"), bootImage.InstallRepo)
		if viper.GetBool("noverifyssl") {
			commandLine += " rd.noverifyssl noverifyssl inst.noverifyssl"
		}
	} else if len(bootImage.LiveImage) > 0 && !strings.Contains(bootImage.CommandLine, "live") {
		commandLine += fmt.Sprintf(" root=live:%s/_/file/liveimg?token=%s", baseURI, c.QueryParam("token"))
		if viper.GetBool("noverifyssl") {
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

	log.Infof("FILE Got valid boot claims: %v", claims)

	macStr := claims.MAC
	if macStr == "" {
		log.WithFields(logrus.Fields{
			"ip": c.RealIP(),
		}).Error("Bad request missing MAC address")
		return echo.NewHTTPError(http.StatusBadRequest, "missing MAC address parameter")
	}

	mac, err := net.ParseMAC(macStr)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":  c.RealIP(),
			"mac": macStr,
			"err": err,
		}).Error("Bad request invalid MAC address")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid MAC address")
	}

	bootImage, err := h.DB.GetBootImage(claims.ID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":  c.RealIP(),
			"mac": mac.String(),
			"err": err,
		}).Error("Failed to find bootspec for host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid boot spec")
	}

	_, fileType := path.Split(c.Request().URL.Path)

	log.Infof("Got request for file %q to %s", fileType, c.RealIP())

	switch {
	case fileType == "kernel":
		return c.File(bootImage.KernelPath)

	case fileType == "liveimg":
		return c.File(bootImage.LiveImage)

	case fileType == "rootfs":
		return c.File(bootImage.RootFS)

	case strings.HasPrefix(fileType, "initrd-"):
		i, err := strconv.Atoi(fileType[7:])
		if err != nil || i < 0 || i >= len(bootImage.InitrdPaths) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("no initrd with ID %q", i))
		}
		return c.File(bootImage.InitrdPaths[i])
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

	bootImage, err := h.DB.GetBootImage(claims.ID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("Kickstart failed to find boot image for host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid boot spec")
	}

	host, err := h.DB.GetHost(claims.ID)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"host_id": claims.ID,
			"mac":     claims.MAC,
			"err":     err,
		}).Error("Kickstart failed to find host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid host")
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
		"rootpw":    viper.GetString("root_password"),
	}

	return c.Render(http.StatusOK, "kickstart.tmpl", data)
}
