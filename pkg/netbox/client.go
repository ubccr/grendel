// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package netbox

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (c *Client) netBoxApiCall(path string, data io.Reader) (*http.Response, error) {
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

	return c.httpClient.Do(req)
}

func (c *Client) NetBoxFetchConfig(name string) (*http.Response, error) {
	var jsonStr = []byte(`{"query": "query {device_list(filters:{name: {exact: \"` + name + `\"}}) {id}}"}`)

	idRes, err := c.netBoxApiCall("/graphql/", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.WithFields(logrus.Fields{
			"name": name,
			"err":  err,
		}).Error("failed to fetch netbox id")
		return nil, err
	}
	defer idRes.Body.Close()

	if idRes.StatusCode != 200 {
		log.WithFields(logrus.Fields{
			"name": name,
			"code": idRes.StatusCode,
		}).Error("failed to fetch netbox id")
		return nil, ErrBadHttpStatus
	}

	rawJson, err := ioutil.ReadAll(idRes.Body)
	if err != nil {
		log.WithFields(logrus.Fields{
			"name": name,
			"err":  err,
		}).Error("failed to read netbox json")
		return nil, err
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
			"name": name,
			"err":  err,
		}).Error("failed to parse netbox json")
		return nil, err
	}

	if len(data.Data.DeviceList) == 0 {
		log.WithFields(logrus.Fields{
			"name": name,
		}).Warn("No device found in netbox")
		return nil, ErrNotFound
	}

	log.WithFields(logrus.Fields{
		"node_name": name,
		"netbox_id": data.Data.DeviceList[0].ID,
	}).Info("Found device in NetBox")

	nid := data.Data.DeviceList[0].ID

	return c.netBoxApiCall("/api/dcim/devices/"+nid+"/render-config/", nil)
}
