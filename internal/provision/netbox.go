// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func (h *Handler) netBoxFetch(path string, data io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", viper.GetString("provision.netbox_url")+path, data)
	if err != nil {
		return nil, err
	}
	if data != nil {
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Accept", "text/plain")

	}

	req.Header.Set("Authorization", "Token "+viper.GetString("provision.netbox_token"))

	return h.netBoxClient.Do(req)
}

func (h *Handler) NetBoxRenderConfig(c echo.Context) error {
	_, host, _, _, err := h.verifyClaims(c)
	if err != nil {
		return err
	}

	var jsonStr = []byte(`{"query": "query {device_list(filters:{name: {exact: \"` + host.Name + `\"}}) {id}}"}`)

	idRes, err := h.netBoxFetch("/graphql/", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.WithFields(logrus.Fields{
			"name": host.Name,
			"err":  err,
		}).Error("failed to fetch netbox id")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render config").SetInternal(err)
	}
	defer idRes.Body.Close()

	if idRes.StatusCode != 200 {
		log.WithFields(logrus.Fields{
			"name": host.Name,
			"code": idRes.StatusCode,
		}).Error("failed to fetch netbox id")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render config").SetInternal(err)
	}

	rawJson, err := ioutil.ReadAll(idRes.Body)
	if err != nil {
		log.WithFields(logrus.Fields{
			"name": host.Name,
			"err":  err,
		}).Error("failed to read netbox json")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render config").SetInternal(err)
	}

	// {"data": {"device_list": [{"id": "5"}]}}
	type netboxResultItem struct {
		ID string `json:"id"`
	}
	type netboxResult struct {
		Data struct {
			DeviceList []*netboxResultItem `json:"device_list"`
		}
	}

	var data netboxResult
	if err := json.Unmarshal(rawJson, &data); err != nil {
		log.WithFields(logrus.Fields{
			"name": host.Name,
			"err":  err,
		}).Error("failed to parse netbox json")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render config").SetInternal(err)
	}

	if len(data.Data.DeviceList) == 0 {
		log.WithFields(logrus.Fields{
			"name": host.Name,
		}).Error("no devices found in netbox")
		return echo.NewHTTPError(http.StatusNotFound, "")
	}

	log.WithFields(logrus.Fields{
		"name": host.Name,
		"id":   data.Data.DeviceList[0].ID,
	}).Info("netbox id")

	nid := data.Data.DeviceList[0].ID

	res, err := h.netBoxFetch("/api/dcim/devices/"+nid+"/render-config/", nil)
	if err != nil {
		log.WithFields(logrus.Fields{
			"name":      host.Name,
			"netbox_id": nid,
			"err":       err,
		}).Error("failed to fetch rendered netbox config")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render config").SetInternal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.WithFields(logrus.Fields{
			"name":      host.Name,
			"netbox_id": nid,
			"code":      res.StatusCode,
		}).Error("failed to fetch rendered netbox config wrong http code")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render config").SetInternal(err)
	}

	return c.Stream(http.StatusOK, "text/plain", res.Body)
}
