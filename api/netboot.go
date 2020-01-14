package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/bmc"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

type NetbootParams struct {
	Username string           `json:"bmc-user"`
	Password string           `json:"bmc-pass"`
	IPMI     bool             `json:"ipmi"`
	Reboot   bool             `json:"reboot"`
	Nodeset  *nodeset.NodeSet `json:"nodeset"`
	Delay    int              `json:"delay"`
}

func (h *Handler) HostNetBoot(c echo.Context) error {
	params := new(NetbootParams)

	if err := c.Bind(params); err != nil {
		log.WithFields(logrus.Fields{
			"ip":  c.RealIP(),
			"err": err,
		}).Warn("Netboot nodes malformed data")
		return echo.NewHTTPError(http.StatusBadRequest, "malformed input data")
	}

	log.Debugf("Got nodeset: %s", params.Nodeset.String())

	hostList, err := h.DB.Find(params.Nodeset)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"err":     err,
			"nodeset": params.Nodeset.String(),
		}).Error("Failed to find host list from datastore")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find hosts")
	}

	if params.Username == "" {
		params.Username = viper.GetString("bmc_user")
	}
	if params.Password == "" {
		params.Password = viper.GetString("bmc_pass")
	}

	if params.Username == "" || params.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "bmc username and password required")
	}

	result := make(map[string]string, 0)

	for _, h := range hostList {
		err := netBoot(h, params)
		if err != nil {
			log.WithFields(logrus.Fields{
				"err":      err,
				"hostName": h.Name,
				"hostID":   h.ID,
			}).Error("Failed to enable PXE on next boot")
			result[h.Name] = fmt.Sprint("ERROR: %s", err)
		} else {
			result[h.Name] = "OK"
		}

		time.Sleep(time.Duration(params.Delay) * time.Second)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HostPowerStatus(c echo.Context) error {
	params := new(NetbootParams)

	if err := c.Bind(params); err != nil {
		log.WithFields(logrus.Fields{
			"ip":  c.RealIP(),
			"err": err,
		}).Warn("PowerInfo nodes malformed data")
		return echo.NewHTTPError(http.StatusBadRequest, "malformed input data")
	}

	log.Debugf("Got nodeset: %s", params.Nodeset.String())

	hostList, err := h.DB.Find(params.Nodeset)
	if err != nil {
		log.WithFields(logrus.Fields{
			"ip":      c.RealIP(),
			"err":     err,
			"nodeset": params.Nodeset.String(),
		}).Error("Failed to find host list from datastore")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to find hosts")
	}

	if params.Username == "" {
		params.Username = viper.GetString("bmc_user")
	}
	if params.Password == "" {
		params.Password = viper.GetString("bmc_pass")
	}

	if params.Username == "" || params.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "bmc username and password required")
	}

	result := make(map[string]*bmc.System, 0)

	for _, h := range hostList {
		var system *bmc.System

		sysmgr, err := systemMgr(h, params)
		if err != nil {
			log.WithFields(logrus.Fields{
				"err":      err,
				"hostName": h.Name,
				"hostID":   h.ID,
			}).Error("Failed to connect to BMC")
			continue
		}
		defer sysmgr.Logout()

		system, err = sysmgr.GetSystem()
		if err != nil {
			log.WithFields(logrus.Fields{
				"err":      err,
				"hostName": h.Name,
				"hostID":   h.ID,
			}).Error("Failed to fetch system info from BMC")
		}

		result[h.Name] = system
		time.Sleep(time.Duration(params.Delay) * time.Second)
	}

	return c.JSON(http.StatusOK, result)
}

func netBoot(host *model.Host, params *NetbootParams) error {
	sysmgr, err := systemMgr(host, params)
	if err != nil {
		return err
	}
	defer sysmgr.Logout()

	err = sysmgr.EnablePXE()
	if err != nil {
		return err
	}

	if params.Reboot {
		err = sysmgr.PowerCycle()
		if err != nil {
			return err
		}
	}

	return nil
}

func systemMgr(host *model.Host, params *NetbootParams) (bmc.SystemManager, error) {
	bmcIntf := host.InterfaceBMC()
	if bmcIntf == nil {
		return nil, errors.New("BMC interface not found")
	}

	bmcAddress := bmcIntf.FQDN
	if bmcAddress == "" {
		bmcAddress = bmcIntf.IP.String()
	}

	if bmcAddress == "" {
		return nil, errors.New("BMC address not set")
	}

	if params.IPMI {
		ipmi, err := bmc.NewIPMI(bmcAddress, params.Username, params.Password, 623)
		if err != nil {
			return nil, err
		}

		return ipmi, nil
	}

	redfish, err := bmc.NewRedfish(fmt.Sprintf("https://%s", bmcAddress), params.Username, params.Password, true)
	if err != nil {
		return nil, err
	}

	return redfish, nil
}
