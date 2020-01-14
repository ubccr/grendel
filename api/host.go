package api

import (
	"net/http"
	"path"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

func (h *Handler) HostAdd(c echo.Context) error {
	host := new(model.Host)

	if err := c.Bind(host); err != nil {
		log.WithFields(logrus.Fields{
			"ip":  c.RealIP(),
			"err": err,
		}).Warn("Add host malformed data")
		return echo.NewHTTPError(http.StatusBadRequest, "malformed input data")
	}

	log.Debugf("Got host: %#v", host)

	err := c.Validate(host)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":  c.RealIP(),
			"err": err,
		}).Warn("Add host invalid data")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data")
	}

	err = h.DB.SaveHost(host)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":  c.RealIP(),
			"err": err,
		}).Error("Failed to save host to datastore")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save host")
	}

	return c.JSON(http.StatusCreated, host)
}

func (h *Handler) HostList(c echo.Context) error {
	hostList, err := h.DB.HostList()
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":  c.RealIP(),
			"err": err,
		}).Error("Failed to fetch host list from datastore")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch hosts")
	}
	return c.JSON(http.StatusOK, hostList)
}

func (h *Handler) HostFind(c echo.Context) error {
	_, nodesetString := path.Split(c.Request().URL.Path)

	nodeset, err := nodeset.NewNodeSet(nodesetString)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Warn("Invalid nodeset")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data")
	}

	log.Infof("Got nodeset: %s", nodeset.String())

	hostList, err := h.DB.Find(nodeset)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to find hosts from datastore")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find hosts")
	}
	return c.JSON(http.StatusOK, hostList)
}
