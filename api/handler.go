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
	return &Handler{DB: db}, nil
}

func (h *Handler) SetupRoutes(e *echo.Echo) {
	e.GET("/", h.Index).Name = "index"
	r := e.Group("/_/")

	config := middleware.JWTConfig{
		Claims:      &model.BootClaims{},
		ContextKey:  ContextKeyJWT,
		SigningKey:  []byte(viper.GetString("secret")),
		TokenLookup: "query:token",
	}
	r.Use(middleware.JWTWithConfig(config))
	r.GET("ipxe", h.Ipxe)
	r.GET("file/kernel", h.File)
	r.GET("file/liveimg", h.File)
	r.GET("file/initrd-*", h.File)
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

	log.Infof("iPXE Got valid boot claims: %v", claims)

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

	bootImage, err := h.DB.GetBootImage(mac.String())
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":  c.RealIP(),
			"mac": mac.String(),
			"err": err,
		}).Error("Failed to find bootspec for host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid boot spec")
	}

	baseURI := fmt.Sprintf("%s://%s", c.Scheme(), c.Request().Host)

	if len(bootImage.LiveImage) > 0 && !strings.Contains(bootImage.CommandLine, "live") {
		bootImage.CommandLine += fmt.Sprintf(" rd.noverifyssl root=live:%s/_/file/liveimg?token=%s", baseURI, c.QueryParam("token"))
	}

	data := map[string]interface{}{
		"token":     c.QueryParam("token"),
		"bootimage": bootImage,
		"mac":       mac,
		"baseuri":   baseURI,
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

	bootImage, err := h.DB.GetBootImage(mac.String())
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
