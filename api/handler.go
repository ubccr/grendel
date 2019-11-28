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
		ContextKey:  "bootspec",
		SigningKey:  []byte("secret"), // TODO: obviously fix this
		TokenLookup: "query:token",
	}
	r.Use(middleware.JWTWithConfig(config))
	r.GET("ipxe", h.Ipxe)
	r.GET("file/kernel", h.File)
	r.GET("file/initrd-*", h.File)
}

func (h *Handler) Index(c echo.Context) error {
	resp := map[string]interface{}{
		"status": "up",
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) Ipxe(c echo.Context) error {
	bootToken := c.Get("bootspec").(*jwt.Token)
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

	bootSpec, err := h.DB.GetBootSpec(mac.String())
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":  c.RealIP(),
			"mac": mac.String(),
			"err": err,
		}).Error("Failed to find bootspec for host")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid boot spec")
	}

	data := map[string]interface{}{
		"token":    c.QueryParam("token"),
		"bootspec": bootSpec,
		"mac":      mac,
		"baseuri":  fmt.Sprintf("%s://%s", c.Scheme(), c.Request().Host),
	}

	return c.Render(http.StatusOK, "ipxe.tmpl", data)
}

func (h *Handler) File(c echo.Context) error {
	bootToken := c.Get("bootspec").(*jwt.Token)
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

	bootSpec, err := h.DB.GetBootSpec(mac.String())
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
		return h.serveBlob(c, fileType, bootSpec.Kernel)

	case strings.HasPrefix(fileType, "initrd-"):
		i, err := strconv.Atoi(fileType[7:])
		if err != nil || i < 0 || i >= len(bootSpec.Initrd) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("no initrd with ID %q", i))
		}
		return h.serveBlob(c, fileType, bootSpec.Initrd[i])
	}

	return echo.NewHTTPError(http.StatusNotFound, "")
}

func (h *Handler) serveBlob(c echo.Context, name string, data []byte) error {
	http.ServeContent(c.Response(), c.Request(), name, time.Time{}, bytes.NewReader(data))
	return nil
}
