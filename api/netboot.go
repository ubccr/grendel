package api

import (
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
		result[h.Name] = netBoot(h, params)
		time.Sleep(time.Duration(params.Delay) * time.Second)
	}

	return c.JSON(http.StatusOK, result)
}

func netBoot(host *model.Host, params *NetbootParams) string {
	bmcIntf := host.InterfaceBMC()
	if bmcIntf == nil {
		return "ERROR: BMC interface not found"
	}

	bmcAddress := bmcIntf.FQDN
	if bmcAddress == "" {
		bmcAddress = bmcIntf.IP.String()
	}

	if bmcAddress == "" {
		return "ERROR: BMC address not set"
	}

	var sysmgr bmc.SystemManager
	var err error

	if params.IPMI {
		sysmgr, err = bmc.NewIPMI(bmcAddress, params.Username, params.Password, 623)
		if err != nil {
			log.WithFields(logrus.Fields{
				"err":      err,
				"hostName": host.Name,
				"hostID":   host.ID,
				"bmc":      bmcAddress,
			}).Error("Failed to create new IPMI system manager")
			return "ERROR: Failed to connect to IPMI"
		}
	} else {
		redfish, err := bmc.NewRedfish(fmt.Sprintf("https://%s", bmcAddress), params.Username, params.Password, true)
		if err != nil {
			log.WithFields(logrus.Fields{
				"err":      err,
				"hostName": host.Name,
				"hostID":   host.ID,
				"bmc":      bmcAddress,
			}).Error("Failed to create new redfish system manager")
			return "ERROR: Failed to connect to redfish"
		}
		defer redfish.Logout()
		sysmgr = redfish
	}

	err = sysmgr.EnablePXE()
	if err != nil {
		log.WithFields(logrus.Fields{
			"err":      err,
			"hostName": host.Name,
			"hostID":   host.ID,
			"bmc":      bmcAddress,
		}).Error("Failed to enable PXE on next boot")
		return "ERROR: Failed to enable PXE on next boot"
	}

	if params.Reboot {
		err = sysmgr.PowerCycle()
		if err != nil {
			log.WithFields(logrus.Fields{
				"err":      err,
				"hostName": host.Name,
				"hostID":   host.ID,
				"bmc":      bmcAddress,
			}).Error("Failed to reboot node")
			return "ERROR: Failed to reboot node"
		}
	}

	return "OK"
}
